package vsphere

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type VSphere struct {
	Server    string `toml:"server"`
	Username  string `toml:"username"`
	Password  string `toml:"password"`
	Insecure  bool   `toml:"insecure"`
	objectMap *ObjectMap
	Summary   *Summary
	client    *govmomi.Client
}

type Summary struct {
	respoolSummary      map[types.ManagedObjectReference]map[string]string
	respoolExtraMetrics map[types.ManagedObjectReference]map[string]int64
	hostSummary         map[types.ManagedObjectReference]map[string]string
	hostExtraMetrics    map[types.ManagedObjectReference]map[string]int64
	vmSummary           map[types.ManagedObjectReference]map[string]string
	vmExtraMetrics      map[types.ManagedObjectReference]map[string]int64
	dsSummary           map[types.ManagedObjectReference]map[string]string
	dsExtraMetrics      map[types.ManagedObjectReference]map[string]int64
}

type ObjectMap struct {
	vmToPool      map[types.ManagedObjectReference]string
	vmToCluster   map[types.ManagedObjectReference]string
	hostToCluster map[types.ManagedObjectReference]string
	morToName     map[types.ManagedObjectReference]string
	metricToName  map[int32]string
}

type MetricDef struct {
	Metric    string
	Instances string
	Key       int32
}

type Metric struct {
	ObjectType []string
	Definition []MetricDef
}

type MetricGroup struct {
	ObjectType string
	Metrics    []MetricDef
}

type EntityQuery struct {
	Name    string
	Entity  types.ManagedObjectReference
	Metrics []int32
}

var (
	sampleConfig = `
  ## FQDN or an IP of a vCenter Server or ESXi host
  server = "vcenter.domain.com"
  ## A vSphere/ESX user
  ## must have System.View and Performance.ModifyIntervals privileges
  username = "root"
  ## Password
  password = "vmware"
  ## Do not validate server's TLS certificate
  # insecure =  true
  ## Host name patterns
  # hosts_system_metrics = ["*"]
  ## Datastore name patterns
  # virtualmachine_metrics = ["*"]
  ## Virtual machine name patterns
  # datastore_metrics = ["*"]
`
	vmfsRegexp = regexp.MustCompile(".*/(.*)/$")
)

func (v *VSphere) Description() string {
	return "Collect metrics from VMware vSphere"
}

func (v *VSphere) SampleConfig() string {
	return sampleConfig
}

func (v *VSphere) Connect() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	u, err := url.Parse("https://" + v.Username + ":" + v.Password + "@" + v.Server + "/sdk")
	if err != nil {
		return err
	}

	client, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		return err
	}

	v.client = client

	return nil
}

// Disconnect from the vCenter
func (v *VSphere) Disconnect() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if v.client != nil {
		if err := v.client.Logout(ctx); err != nil {
			return err
		}
	}

	return nil
}

// SetMetricToName SetMetricToName
func (v *VSphere) SetMetricToName() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := v.client

	var perfmanager mo.PerformanceManager
	if err := client.RetrieveOne(ctx, *client.ServiceContent.PerfManager, nil, &perfmanager); err != nil {
		return err
	}

	v.objectMap.metricToName = make(map[int32]string)
	for _, perf := range perfmanager.PerfCounter {
		groupinfo := perf.GroupInfo.GetElementDescription()
		nameinfo := perf.NameInfo.GetElementDescription()
		identifier := groupinfo.Key + "_" + nameinfo.Key + "_" + fmt.Sprint(perf.RollupType)
		v.objectMap.metricToName[perf.Key] = identifier
	}
	return nil
}

func (v *VSphere) getDatacenters(ctx context.Context) ([]types.ManagedObjectReference, error) {
	var rootFolder mo.Folder
	err := v.client.RetrieveOne(ctx, v.client.ServiceContent.RootFolder, nil, &rootFolder)
	if err != nil {
		return nil, err
	}

	datacenters := []types.ManagedObjectReference{}
	for _, child := range rootFolder.ChildEntity {
		datacenters = append(datacenters, child)
	}
	return datacenters, nil
}

func (v *VSphere) getAllManagedObjectReference(ctx context.Context, datacenters []types.ManagedObjectReference) ([]types.ManagedObjectReference, error) {

	var viewManager mo.ViewManager
	err := v.client.RetrieveOne(ctx, *v.client.ServiceContent.ViewManager, nil, &viewManager)
	if err != nil {
		return nil, err
	}

	mors := []types.ManagedObjectReference{}
	for _, datacenter := range datacenters {
		req := types.CreateContainerView{
			This:      viewManager.Reference(),
			Container: datacenter,
			Recursive: true}

		res, err := methods.CreateContainerView(ctx, v.client.RoundTripper, &req)
		if err != nil {
			return nil, err
		}
		// Retrieve the created ContentView
		var containerView mo.ContainerView
		err = v.client.RetrieveOne(ctx, res.Returnval, nil, &containerView)
		if err != nil {
			return nil, err
		}
		mors = append(mors, containerView.View...)
	}

	return mors, nil
}

func (v *VSphere) getAllPerformances(ctx context.Context, mors []types.ManagedObjectReference) (*types.QueryPerfResponse, error) {
	queries := []types.PerfQuerySpec{}

	// Common parameters
	intervalIDint := 20
	var intervalID int32
	intervalID = int32(intervalIDint)

	endTime := time.Now().Add(time.Duration(-1) * time.Second)
	startTime := endTime.Add(time.Duration(-60) * time.Second)

	// Parse objects
	for _, mor := range mors {

		perMetrics := types.QueryAvailablePerfMetric{This: *v.client.ServiceContent.PerfManager, Entity: mor, BeginTime: &startTime, EndTime: &endTime, IntervalId: intervalID}
		perfres, err := methods.QueryAvailablePerfMetric(ctx, v.client.RoundTripper, &perMetrics)
		if err != nil {
			continue
		}

		metricIds := []types.PerfMetricId{}
		for _, perf := range perfres.Returnval {
			metricIds = append(metricIds, types.PerfMetricId{CounterId: perf.CounterId, Instance: "*"})
		}

		if len(metricIds) > 0 {
			queries = append(queries, types.PerfQuerySpec{Entity: mor, StartTime: &startTime, EndTime: &endTime, MetricId: metricIds, IntervalId: intervalID})
		}

	}

	// Query the performances
	perfreq := types.QueryPerf{This: *v.client.ServiceContent.PerfManager, QuerySpec: queries}
	perfres, err := methods.QueryPerf(ctx, v.client.RoundTripper, &perfreq)
	if err != nil {
		return nil, err
	}
	return perfres, nil
}

func (v *VSphere) setSummaryObjectMap(ctx context.Context, mors []types.ManagedObjectReference) error {
	// Create MORS for each object type
	vmRefs := []types.ManagedObjectReference{}
	hostRefs := []types.ManagedObjectReference{}
	clusterRefs := []types.ManagedObjectReference{}
	respoolRefs := []types.ManagedObjectReference{}
	datastoreRefs := []types.ManagedObjectReference{}
	newMors := []types.ManagedObjectReference{}
	for _, mor := range mors {
		switch morType := mor.Type; morType {
		case "VirtualMachine":
			vmRefs = append(vmRefs, mor)
			newMors = append(newMors, mor)
		case "HostSystem":
			hostRefs = append(hostRefs, mor)
			newMors = append(newMors, mor)
		case "ClusterComputeResource":
			clusterRefs = append(clusterRefs, mor)
		case "ResourcePool":
			respoolRefs = append(respoolRefs, mor)
		case "Datastore":
			datastoreRefs = append(datastoreRefs, mor)
		default:
		}
	}

	pc := property.DefaultCollector(v.client.Client)
	// Retrieve properties for all vms
	var vmmo []mo.VirtualMachine
	if len(vmRefs) > 0 {
		if err := pc.Retrieve(ctx, vmRefs, []string{"summary", "config"}, &vmmo); err != nil {
			return err
		}
	}

	// Retrieve properties for hosts
	var hsmo []mo.HostSystem
	if len(hostRefs) > 0 {
		if err := pc.Retrieve(ctx, hostRefs, []string{"parent", "summary", "hardware"}, &hsmo); err != nil {
			return err
		}
	}

	//Retrieve properties for Cluster(s)
	var clmo []mo.ClusterComputeResource
	if len(clusterRefs) > 0 {
		if err := pc.Retrieve(ctx, clusterRefs, []string{"name", "configuration", "host"}, &clmo); err != nil {
			return err
		}
	}

	//Retrieve properties for ResourcePool
	var rpmo []mo.ResourcePool
	if len(respoolRefs) > 0 {
		if err := pc.Retrieve(ctx, respoolRefs, []string{"summary", "name", "config", "vm"}, &rpmo); err != nil {
			return err
		}
	}

	// Retrieve summary property for all datastores
	var dss []mo.Datastore
	if len(datastoreRefs) > 0 {
		if err := pc.Retrieve(ctx, datastoreRefs, []string{"summary", "info"}, &dss); err != nil {
			return err
		}
	}

	v.objectMap.hostToCluster = make(map[types.ManagedObjectReference]string)
	v.objectMap.vmToCluster = make(map[types.ManagedObjectReference]string)
	v.Summary.vmSummary = make(map[types.ManagedObjectReference]map[string]string)
	v.Summary.vmExtraMetrics = make(map[types.ManagedObjectReference]map[string]int64)
	for _, vm := range vmmo {
		// check if VM is a clone in progress and skip it
		if vm.Summary.Runtime.Host == nil {
			continue
		}
		vmhost := vm.Summary.Runtime.Host

		for _, cl := range clmo {
			for _, host := range cl.Host {
				v.objectMap.hostToCluster[host] = cl.Name

				if *vmhost == host {
					v.objectMap.vmToCluster[vm.Self] = cl.Name
				}
			}
		}
	}

	v.Summary.hostSummary = make(map[types.ManagedObjectReference]map[string]string)
	v.Summary.hostExtraMetrics = make(map[types.ManagedObjectReference]map[string]int64)
	for _, host := range hsmo {
		// Extra tags per host
		v.Summary.hostSummary[host.Self] = make(map[string]string)
		v.Summary.hostSummary[host.Self]["name"] = host.Summary.Config.Name
		v.Summary.hostSummary[host.Self]["host_uuid"] = host.Hardware.SystemInfo.Uuid
		v.Summary.hostSummary[host.Self]["cluster"] = v.objectMap.hostToCluster[host.Self]

		// Extra metrics per host
		v.Summary.hostExtraMetrics[host.Self] = make(map[string]int64)
		v.Summary.hostExtraMetrics[host.Self]["uptime"] = int64(host.Summary.QuickStats.Uptime)
		v.Summary.hostExtraMetrics[host.Self]["cpu_corecount_total"] = int64(host.Summary.Hardware.NumCpuThreads)
	}

	// Retrieve properties for the pools
	v.Summary.respoolSummary = make(map[types.ManagedObjectReference]map[string]string)
	v.Summary.respoolExtraMetrics = make(map[types.ManagedObjectReference]map[string]int64)
	v.objectMap.vmToPool = make(map[types.ManagedObjectReference]string)
	for _, pool := range rpmo {
		v.Summary.respoolSummary[pool.Self] = make(map[string]string)
		v.Summary.respoolSummary[pool.Self]["name"] = pool.Name
		v.Summary.respoolExtraMetrics[pool.Self] = make(map[string]int64)
		v.Summary.respoolExtraMetrics[pool.Self]["cpu_limit"] = *pool.Config.CpuAllocation.Limit
		v.Summary.respoolExtraMetrics[pool.Self]["memory_limit"] = *pool.Config.MemoryAllocation.Limit

		for _, vm := range pool.Vm {
			v.objectMap.vmToPool[vm] = pool.Name
		}
	}

	// Initialize the maps that will hold the extra tags and metrics for VMs
	v.Summary.vmSummary = make(map[types.ManagedObjectReference]map[string]string)
	v.Summary.vmExtraMetrics = make(map[types.ManagedObjectReference]map[string]int64)
	for _, vm := range vmmo {
		v.Summary.vmSummary[vm.Self] = make(map[string]string)
		// Ugly way to extract datastore value
		re, _ := regexp.Compile(`\[(.*?)\]`)
		v.Summary.vmSummary[vm.Self]["datastore"] = strings.Replace(strings.Replace(re.FindString(fmt.Sprintln(vm.Summary.Config.VmPathName)), "[", "", -1), "]", "", -1)
		if v.objectMap.vmToCluster[vm.Self] != "" {
			v.Summary.vmSummary[vm.Self]["cluster"] = v.objectMap.vmToCluster[vm.Self]
		}
		if v.objectMap.vmToPool[vm.Self] != "" {
			v.Summary.vmSummary[vm.Self]["respool"] = v.objectMap.vmToPool[vm.Self]
		}
		if vm.Summary.Runtime.Host != nil {
			v.Summary.vmSummary[vm.Self]["esx"] = v.Summary.hostSummary[*vm.Summary.Runtime.Host]["host_uuid"]
		}
		v.Summary.vmSummary[vm.Self]["domain_id"] = vm.Config.Uuid

		// Extra metrics per VM
		v.Summary.vmExtraMetrics[vm.Self] = make(map[string]int64)
		v.Summary.vmExtraMetrics[vm.Self]["uptime"] = int64(vm.Summary.QuickStats.UptimeSeconds)
	}

	// get object names
	objects := []mo.ManagedEntity{}

	//object for propery collection
	propSpec := &types.PropertySpec{Type: "ManagedEntity", PathSet: []string{"name"}}
	var objectSet []types.ObjectSpec
	for _, mor := range mors {
		objectSet = append(objectSet, types.ObjectSpec{Obj: mor, Skip: types.NewBool(false)})
	}

	//retrieve name property
	propreq := types.RetrieveProperties{SpecSet: []types.PropertyFilterSpec{{ObjectSet: objectSet, PropSet: []types.PropertySpec{*propSpec}}}}
	propres, err := v.client.PropertyCollector().RetrieveProperties(ctx, propreq)
	if err != nil {
		return err
	}

	//load retrieved properties
	err = mo.LoadRetrievePropertiesResponse(propres, &objects)
	if err != nil {
		return err
	}

	//create a map to resolve object names
	v.objectMap.morToName = make(map[types.ManagedObjectReference]string)
	for _, object := range objects {
		v.objectMap.morToName[object.Self] = object.Name
	}

	v.Summary.dsSummary = make(map[types.ManagedObjectReference]map[string]string)
	v.Summary.dsExtraMetrics = make(map[types.ManagedObjectReference]map[string]int64)
	for _, datastore := range dss {
		v.Summary.dsSummary[datastore.Self] = make(map[string]string)
		v.Summary.dsSummary[datastore.Self]["name"] = datastore.Summary.Name
		v.Summary.dsSummary[datastore.Self]["uuid"] = ""
		if str := vmfsRegexp.FindStringSubmatch(datastore.Info.GetDatastoreInfo().Url); len(str) > 1 {
			v.Summary.dsSummary[datastore.Self]["uuid"] = str[1]
		}

		v.Summary.dsExtraMetrics[datastore.Self] = make(map[string]int64)
		v.Summary.dsExtraMetrics[datastore.Self]["capacity"] = datastore.Summary.Capacity
		v.Summary.dsExtraMetrics[datastore.Self]["free_space"] = datastore.Summary.FreeSpace

	}
	return nil
}

func (v *VSphere) QueryPerformances() (*types.QueryPerfResponse, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dcs, err := v.getDatacenters(ctx)
	if err != nil {
		return nil, err
	}

	mors, err := v.getAllManagedObjectReference(ctx, dcs)
	if err != nil {
		return nil, err
	}

	if err := v.setSummaryObjectMap(ctx, mors); err != nil {
		return nil, err
	}

	performances, err := v.getAllPerformances(ctx, mors)
	if err != nil {
		return nil, err
	}

	return performances, nil
}

func (v *VSphere) getFieldTag(base types.BasePerfEntityMetricBase) (map[string]interface{}, map[string]string) {
	pem := base.(*types.PerfEntityMetric)
	name := strings.ToLower(v.objectMap.morToName[pem.Entity])
	fields := make(map[string]interface{})
	tags := map[string]string{"name": name}

	if summary, ok := v.Summary.vmSummary[pem.Entity]; ok {
		for key, tag := range summary {
			tags[key] = tag
		}
	}
	if summary, ok := v.Summary.hostSummary[pem.Entity]; ok {
		for key, tag := range summary {
			tags[key] = tag
		}
	}

	if summary, ok := v.Summary.respoolSummary[pem.Entity]; ok {
		for key, tag := range summary {
			tags[key] = tag
		}
	}

	// Create the fields for the hostExtraMetrics
	if metrics, ok := v.Summary.hostExtraMetrics[pem.Entity]; ok {
		for key, value := range metrics {
			fields[key] = value
		}
	}
	// Create the fields for the vmExtraMetrics
	if metrics, ok := v.Summary.vmExtraMetrics[pem.Entity]; ok {
		for key, value := range metrics {
			fields[key] = value
		}
	}
	return fields, tags
}

func (v *VSphere) SetAcc(acc telegraf.Accumulator, performances *types.QueryPerfResponse) {
	nowTime := time.Now()
	for _, base := range performances.Returnval {
		pem := base.(*types.PerfEntityMetric)
		fields, tags := v.getFieldTag(base)

		entityName := strings.ToLower(pem.Entity.Type)

		specialFields := make(map[string]map[string]map[string]map[string]interface{})
		specialTags := make(map[string]map[string]map[string]map[string]string)
		for _, baseserie := range pem.Value {
			serie := baseserie.(*types.PerfMetricIntSeries)
			metricName := strings.ToLower(v.objectMap.metricToName[serie.Id.CounterId])
			instanceName := strings.ToLower(strings.Replace(serie.Id.Instance, ".", "_", -1))
			measurementName := strings.Split(metricName, "_")[0]

			if strings.Index(metricName, "datastore") != -1 {
				instanceName = ""
			}

			var value int64 = -1
			if strings.HasSuffix(metricName, "_average") {
				value = average(serie.Value...)
			} else if strings.HasSuffix(metricName, "_maximum") {
				value = max(serie.Value...)
			} else if strings.HasSuffix(metricName, "_minimum") {
				value = min(serie.Value...)
			} else if strings.HasSuffix(metricName, "_latest") {
				value = serie.Value[len(serie.Value)-1]
			} else if strings.HasSuffix(metricName, "_summation") {
				value = sum(serie.Value...)
			} else {
				value = serie.Value[len(serie.Value)-1]
			}

			if instanceName == "" {
				fields[metricName] = value
			} else {
				// init maps
				if specialFields[measurementName] == nil {
					specialFields[measurementName] = make(map[string]map[string]map[string]interface{})
					specialTags[measurementName] = make(map[string]map[string]map[string]string)

				}

				if specialFields[measurementName][tags["name"]] == nil {
					specialFields[measurementName][tags["name"]] = make(map[string]map[string]interface{})
					specialTags[measurementName][tags["name"]] = make(map[string]map[string]string)
				}

				if specialFields[measurementName][tags["name"]][instanceName] == nil {
					specialFields[measurementName][tags["name"]][instanceName] = make(map[string]interface{})
					specialTags[measurementName][tags["name"]][instanceName] = make(map[string]string)

				}

				specialFields[measurementName][tags["name"]][instanceName][metricName] = value

				for k, v := range tags {
					specialTags[measurementName][tags["name"]][instanceName][k] = v
				}
				specialTags[measurementName][tags["name"]][instanceName]["instance"] = instanceName
			}
		}

		acc.AddFields(entityName, fields, tags, nowTime)
		for measurement, v := range specialFields {
			for name, metric := range v {
				for instance, value := range metric {
					acc.AddFields(measurement, value, specialTags[measurement][name][instance], nowTime)
				}
			}
		}
	}

	for p, s := range v.Summary.respoolSummary {
		respoolFields := map[string]interface{}{
			"cpu_limit":    v.Summary.respoolExtraMetrics[p]["cpu_limit"],
			"memory_limit": v.Summary.respoolExtraMetrics[p]["memory_limit"],
		}
		respoolTags := map[string]string{"pool_name": s["name"]}
		acc.AddFields("resourcepool", respoolFields, respoolTags, nowTime)
	}

	for d, s := range v.Summary.dsSummary {
		datastoreFields := map[string]interface{}{
			"capacity":   v.Summary.dsExtraMetrics[d]["capacity"],
			"free_space": v.Summary.dsExtraMetrics[d]["free_space"],
		}
		datastoreTags := map[string]string{"ds_name": s["name"], "uuid": s["uuid"]}
		acc.AddFields("datastore", datastoreFields, datastoreTags, nowTime)
	}
}

func (v *VSphere) Gather(acc telegraf.Accumulator) error {

	var err error

	// Connect and log in to ESX or vCenter
	if err = v.Connect(); err != nil {
		return fmt.Errorf("Failed to connect the server %s: %s", v.Server, err)
	}

	v.objectMap = &ObjectMap{}
	v.Summary = &Summary{}
	if err = v.SetMetricToName(); err != nil {
		return fmt.Errorf("Failed to evaluate metrics %s: %s", v.Server, err)
	}

	performances, err := v.QueryPerformances()
	if err != nil {
		return fmt.Errorf("Failed to query %s: %s", v.Server, err)
	}

	v.SetAcc(acc, performances)

	// spew.Dump(acc)
	if err = v.Disconnect(); err != nil {
		return fmt.Errorf("Failed to disconnect the server %s: %s", v.Server, err)
	}

	return nil
}

func init() {
	inputs.Add("vsphere", func() telegraf.Input { return &VSphere{} })
}

func min(n ...int64) int64 {
	var min int64 = -1
	for _, i := range n {
		if i >= 0 {
			if min == -1 {
				min = i
			} else {
				if i < min {
					min = i
				}
			}
		}
	}
	return min
}

func max(n ...int64) int64 {
	var max int64 = -1
	for _, i := range n {
		if i >= 0 {
			if max == -1 {
				max = i
			} else {
				if i > max {
					max = i
				}
			}
		}
	}
	return max
}

func sum(n ...int64) int64 {
	var total int64
	for _, i := range n {
		if i > 0 {
			total += i
		}
	}
	return total
}

func average(n ...int64) int64 {
	var total int64
	var count int64
	for _, i := range n {
		if i >= 0 {
			count++
			total += i
		}
	}
	favg := float64(total) / float64(count)
	return int64(math.Floor(favg + .5))
}
