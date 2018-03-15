package dcaitype

const (
	AgentUnknown = AgentType(0)
	AgentLinux   = AgentType(1)
	AgentVMware  = AgentType(2)
	AgentWindows = AgentType(3)

	ClusterUnknown        = ClusterType(0)
	ClusterDefaultCluster = ClusterType(1)
	ClustervSAN           = ClusterType(2)
	ClusterKubernetes     = ClusterType(3)
	ClusterCeph           = ClusterType(4)

	OSUnknown = OSType(0)
	OSLinux   = OSType(1)
	OSVMware  = OSType(2)
	OSWindows = OSType(3)

	BlkUnknown    = BlkType(0)
	BlkPD         = BlkType(1)
	BlkMegaRaidVD = BlkType(2)

	DiskTypeUnknown = DiskType(0)
	DiskTypeHDD     = DiskType(1)
	DiskTypeSSD     = DiskType(2)
	DiskTypeSSDNVME = DiskType(3)
	DiskTypeSSDSAS  = DiskType(4)
	DiskTypeSSDSATA = DiskType(5)
	DiskTypeHDDSAS  = DiskType(6)
	DiskTypeHDDSATA = DiskType(7)

	DiskStatusUnknown  = DiskStatusType(0)
	DiskStatusGood     = DiskStatusType(1)
	DiskStatusFailure  = DiskStatusType(2)
	DiskStatusWarning  = DiskStatusType(3)
	DiskStatusCritical = DiskStatusType(4)

	EventTitleUnknown       = EventTitle(0)
	EventTitleAgentStarted  = EventTitle(1)
	EventTitleAgentAlived   = EventTitle(2)
	EventTitleAgentStopped  = EventTitle(3)
	EventTitleSmartDataSent = EventTitle(4)
	EventTitleHostDataSent  = EventTitle(5)

	EventTypeUnknown                = EventType(0)
	EventTypeFirstAgentHeartbeat    = EventType(1)
	EventTypeIntervalAgentHeartbeat = EventType(2)
	EventTypeMetricsMonitoring      = EventType(3)

	LogLevelDebug   = LogLevel(0) // General debugging information: basically useful information that is used for debugging purposes
	LogLevelInfo    = LogLevel(1) // General information: Logs that track the general flow of the application
	LogLevelWarning = LogLevel(2) // Abnormal information: Logs that highlight an abnormal infomation cause by incorrect configuration, temporarily unavailable services or timeout that may be ignored
	LogLevelError   = LogLevel(3) // Handled exceptions: Errors due to program or system errors that need to be solved immediately
	LogLevelFatal   = LogLevel(4) // Unhandled exceptions: critical errors that cause program breakdown
)

var (
	ClusterTypes = map[int]string{
		0: "Unknown Cluster Type",
		1: "NoCluster",
		2: "vSAN",
		3: "Kubernetes",
		4: "ceph",
	}
	AgentTypes = map[int]string{
		0: "Unknown agent type",
		1: "linux",
		2: "vmware",
		3: "windows",
	}
	OsTypes = map[int]string{
		0: "Unknown OS type",
		1: "linux",
		2: "vmware",
		3: "windows",
	}
	BlkTypes = map[int]string{
		0: "Unknown Block Device Type",
		1: "Physical Disk",
		2: "MegaRaid Virtual Disk",
	}
	DiskTypes = map[int]string{
		0: "Unknown",
		1: "HDD",
		2: "SSD",
		3: "SSD NVME",
		4: "SSD SAS",
		5: "SSD SATA",
		6: "HDD SAS",
		7: "HDD SATA",
	}
	DiskStatusTypes = map[int]string{
		0: "Unknown",
		1: "Good",
		2: "Failure",
		3: "Warning",
		4: "Critical",
	}
	EventTitles = map[int]string{
		0: "Unknown",
		1: "The Agent is started",
		2: "The Agent is still alived",
		3: "The Agent is stopped",
		4: "Data of Raw SMART was written to DB",
		5: "Data of Host was written to DB",
	}
	EventTypes = map[int]string{
		0: "Unknown",
		1: "First Agent Heartbeat",
		2: "Interval Agent Heartbeat",
		3: "Metrics Monitoring",
	}
	LogLevels = map[int]string{
		0: "Debug",
		1: "Info",
		2: "Warning",
		3: "Error",
		4: "Fatal",
	}
)

type AgentType int

func (a AgentType) String() string {
	return AgentTypes[int(a)]
}

func (a AgentType) LookupCode(typestr string) AgentType {
	t := lookupType(typestr, a)
	if tt, ok := t.(AgentType); ok {
		return tt
	}
	return AgentType(0)
}

type ClusterType int

func (c ClusterType) String() string {
	return ClusterTypes[int(c)]
}

func (c ClusterType) LookupCode(typestr string) ClusterType {
	t := lookupType(typestr, c)
	if tt, ok := t.(ClusterType); ok {
		return tt
	}
	return ClusterType(0)
}

type OSType int

func (o OSType) String() string {
	return OsTypes[int(o)]
}

func (o OSType) LookupCode(typestr string) OSType {
	t := lookupType(typestr, o)
	if tt, ok := t.(OSType); ok {
		return tt
	}
	return OSType(0)
}

type BlkType int

func (b BlkType) String() string {
	return BlkTypes[int(b)]
}

func (b BlkType) LookupCode(typestr string) BlkType {
	t := lookupType(typestr, b)
	if tt, ok := t.(BlkType); ok {
		return tt
	}
	return BlkType(0)
}

type DiskType int

func (d DiskType) String() string {
	return DiskTypes[int(d)]
}

func (d DiskType) LookupCode(typestr string) DiskType {
	t := lookupType(typestr, d)
	if tt, ok := t.(DiskType); ok {
		return tt
	}
	return DiskType(0)
}

type DiskStatusType int

func (d DiskStatusType) String() string {
	return DiskStatusTypes[int(d)]
}

func (d DiskStatusType) LookupCode(typestr string) DiskStatusType {
	t := lookupType(typestr, d)
	if tt, ok := t.(DiskStatusType); ok {
		return tt
	}
	return DiskStatusType(0)
}

type EventTitle int

func (e EventTitle) String() string {
	return EventTitles[int(e)]
}

func (e EventTitle) LookupCode(typestr string) EventTitle {
	t := lookupType(typestr, e)
	if tt, ok := t.(EventTitle); ok {
		return tt
	}
	return EventTitle(0)
}

type EventType int

func (e EventType) String() string {
	return EventTypes[int(e)]
}

func (e EventType) LookupCode(typestr string) EventType {
	t := lookupType(typestr, e)
	if tt, ok := t.(EventType); ok {
		return tt
	}
	return EventType(0)
}

type LogLevel int

func (l LogLevel) String() string {
	return LogLevels[int(l)]
}

func (l LogLevel) LookupCode(typestr string) LogLevel {
	t := lookupType(typestr, l)
	if tt, ok := t.(LogLevel); ok {
		return tt
	}
	return LogLevel(0)
}

func lookupType(typestr string, t interface{}) interface{} {
	switch t.(type) {
	case AgentType:
		return AgentType(lookupTypeInternal(typestr, AgentTypes))
	case ClusterType:
		return ClusterType(lookupTypeInternal(typestr, ClusterTypes))
	case OSType:
		return OSType(lookupTypeInternal(typestr, OsTypes))
	case BlkType:
		return BlkType(lookupTypeInternal(typestr, BlkTypes))
	case DiskType:
		return DiskType(lookupTypeInternal(typestr, DiskTypes))
	case DiskStatusType:
		return DiskStatusType(lookupTypeInternal(typestr, DiskStatusTypes))
	case EventTitle:
		return EventTitle(lookupTypeInternal(typestr, EventTitles))
	case EventType:
		return EventType(lookupTypeInternal(typestr, EventTypes))
	case LogLevel:
		return LogLevel(lookupTypeInternal(typestr, LogLevels))
	default:
		return nil
	}
}

func lookupTypeInternal(t string, m map[int]string) int {
	for k, v := range m {
		if t == v {
			return k
		}
	}
	return 0
}
