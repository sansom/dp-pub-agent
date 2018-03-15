package dcai

import (
	"fmt"
	"log"

	"github.com/influxdata/telegraf/dcai/connector/vcsa"
	saicluster "github.com/influxdata/telegraf/dcai/sai/cluster"
	"github.com/influxdata/telegraf/dcai/topology/cluster"
	"github.com/influxdata/telegraf/dcai/topology/datacenter"
	"github.com/influxdata/telegraf/dcai/topology/host"
	"github.com/influxdata/telegraf/dcai/topology/host/linux"
	"github.com/influxdata/telegraf/dcai/type"
	"github.com/influxdata/telegraf/dcai/util"
	"github.com/influxdata/telegraf/internal/config"
)

var (
	dcaiagent *DcaiAgent // root for all dcai information
)

type DcaiAgent struct {
	Agenttype      dcaitype.AgentType
	TelegrafConfig *config.Config
	nextVersion    string
	version        string
	commit         string
	branch         string
}

func FetchAgentHostConfig(t dcaitype.AgentType, dmidecodePath string) (host.HostConfig, error) {
	if t == dcaitype.AgentLinux || t == dcaitype.AgentVMware {
		return linux.NewLinuxHostConfig(dmidecodePath)
	} else {
		return nil, fmt.Errorf("Unsupported agent host type %s", t.String())
	}
}

func FetchDefaultDatacenter(telegrafConfig *config.Config) (*datacenter.DatacenterConfig, error) {
	var (
		err error
		at  dcaitype.AgentType
		ct  dcaitype.ClusterType
		cl  *cluster.ClusterConfig
		dc  *datacenter.DatacenterConfig
	)
	at = at.LookupCode(telegrafConfig.Agent.AgentType)
	agenthost, err := FetchAgentHostConfig(at, telegrafConfig.Agent.DmidecodePath)
	if err != nil {
		return nil, err
	}

	ct = ct.LookupCode(telegrafConfig.Tags["cluster_type"])
	if ct == dcaitype.ClusterUnknown {
		ct = dcaitype.ClusterDefaultCluster
	}

	if cl, err = cluster.NewClusterConfig(ct, telegrafConfig.Tags["cluster_name"], telegrafConfig.Tags["cluster_name"]); err != nil {
		return nil, err
	}

	if dc, err = datacenter.NewDatacenterConfig(telegrafConfig.Tags["datacenter_name"], nil, nil); err != nil {
		return nil, err
	}

	cl.AppendHost(agenthost)
	dc.AppendCluster(cl)
	dc.AppendHost(agenthost)

	return dc, nil
}

func FetchVsphereTopology(vcsa *vcsa.VcsaConnector) ([]*datacenter.DatacenterConfig, error) {
	if vcsa == nil {
		return nil, fmt.Errorf("null vcsa connector")
	}

	// connect and defer the disconnect
	err := vcsa.ConnectVsphere()
	if err != nil {
		return nil, fmt.Errorf("%s\n", err)
	}
	defer vcsa.DisconnectVsphere()

	return vcsa.GetTopology()
}

func NewDcaiAgent(config *config.Config, nextver string, ver string, commit string, branch string) (*DcaiAgent, error) {
	if dcaiagent != nil {
		log.Printf("W! dcaiagent instance already exists")
		return dcaiagent, nil
	}

	// translate agent type
	var t dcaitype.AgentType
	t = t.LookupCode(config.Agent.AgentType)
	// default to linux agent type
	if t == dcaitype.AgentUnknown {
		t = dcaitype.AgentLinux
	}

	dcaiagent = new(DcaiAgent)
	dcaiagent.Agenttype = t
	dcaiagent.TelegrafConfig = config
	if config.Agent.DmidecodePath != "" {
		err := util.CheckCmdPath(config.Agent.DmidecodePath)
		if err != nil {
			return nil, fmt.Errorf("Invalid dmidecode_path in config file")
		}
	} else {
		path, err := util.GetCmdPathInOsPath("dmidecode")
		if err != nil {
			return nil, fmt.Errorf("Cannot find dmidecode in system path")
		} else {
			dcaiagent.TelegrafConfig.Agent.DmidecodePath = path
		}
	}
	dcaiagent.nextVersion = nextver
	dcaiagent.version = ver
	dcaiagent.commit = commit
	dcaiagent.branch = branch
	return dcaiagent, nil
}

func GetDcaiAgent() (*DcaiAgent, error) {
	if dcaiagent != nil {
		return dcaiagent, nil
	}
	return nil, fmt.Errorf("dcaiagent instance does not exist")
}

func (a *DcaiAgent) GetTelegrafVersion() string {
	if a.version == "" {
		return fmt.Sprintf("%s-%s", a.nextVersion, a.commit)
	}
	return a.version
}

func (a *DcaiAgent) GetSaiClusterDomainId() string {
	id := a.TelegrafConfig.Agent.SaiClusterDomainId
	if id == "" {
		id = saicluster.SaiClusterDefaultDomainId
	}
	return id
}

func (a *DcaiAgent) GetSaiClusterName() string {
	n := a.TelegrafConfig.Agent.SaiClusterName
	if n == "" {
		n = saicluster.SaiClusterDefaultName
	}
	return n
}
