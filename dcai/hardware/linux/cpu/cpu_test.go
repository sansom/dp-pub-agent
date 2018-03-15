package cpu

import (
	"github.com/influxdata/telegraf/dcai/testutil"
	"testing"
)

var (
	input = `
processor       : 0
vendor_id       : GenuineIntel
cpu family      : 6
model           : 23
model name      : Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz
stepping        : 6
microcode       : 0x60f
cpu MHz         : 2000.000
cache size      : 6144 KB
physical id     : 0
siblings        : 4
core id         : 0
cpu cores       : 4
apicid          : 0
initial apicid  : 0
fpu             : yes
fpu_exception   : yes
cpuid level     : 10
wp              : yes
flags           : fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx lm constant_tsc arch_perfmon pebs bts rep_good nopl aperfmperf pni dtes64 monitor ds_cpl vmx est tm2 ssse3 cx16 xtpr pdcm dca sse4_1 lahf_lm tpr_shadow vnmi flexpriority dtherm
bogomips        : 5985.45
clflush size    : 64
cache_alignment : 64
address sizes   : 38 bits physical, 48 bits virtual
power management:

processor       : 1
vendor_id       : GenuineIntel
cpu family      : 6
model           : 23
model name      : Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz
stepping        : 6
microcode       : 0x60f
cpu MHz         : 2000.000
cache size      : 6144 KB
physical id     : 1
siblings        : 4
core id         : 0
cpu cores       : 4
apicid          : 4
initial apicid  : 4
fpu             : yes
fpu_exception   : yes
cpuid level     : 10
wp              : yes
flags           : fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx lm constant_tsc arch_perfmon pebs bts rep_good nopl aperfmperf pni dtes64 monitor ds_cpl vmx est tm2 ssse3 cx16 xtpr pdcm dca sse4_1 lahf_lm tpr_shadow vnmi flexpriority dtherm
bogomips        : 5985.04
clflush size    : 64
cache_alignment : 64
address sizes   : 38 bits physical, 48 bits virtual
power management:

processor       : 2
vendor_id       : GenuineIntel
cpu family      : 6
model           : 23
model name      : Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz
stepping        : 6
microcode       : 0x60f
cpu MHz         : 2000.000
cache size      : 6144 KB
physical id     : 0
siblings        : 4
core id         : 2
cpu cores       : 4
apicid          : 2
initial apicid  : 2
fpu             : yes
fpu_exception   : yes
cpuid level     : 10
wp              : yes
flags           : fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx lm constant_tsc arch_perfmon pebs bts rep_good nopl aperfmperf pni dtes64 monitor ds_cpl vmx est tm2 ssse3 cx16 xtpr pdcm dca sse4_1 lahf_lm tpr_shadow vnmi flexpriority dtherm
bogomips        : 5985.45
clflush size    : 64
cache_alignment : 64
address sizes   : 38 bits physical, 48 bits virtual
power management:

processor       : 3
vendor_id       : GenuineIntel
cpu family      : 6
model           : 23
model name      : Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz
stepping        : 6
microcode       : 0x60f
cpu MHz         : 2000.000
cache size      : 6144 KB
physical id     : 1
siblings        : 4
core id         : 2
cpu cores       : 4
apicid          : 6
initial apicid  : 6
fpu             : yes
fpu_exception   : yes
cpuid level     : 10
wp              : yes
flags           : fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx lm constant_tsc arch_perfmon pebs bts rep_good nopl aperfmperf pni dtes64 monitor ds_cpl vmx est tm2 ssse3 cx16 xtpr pdcm dca sse4_1 lahf_lm tpr_shadow vnmi flexpriority dtherm
bogomips        : 5985.04
clflush size    : 64
cache_alignment : 64
address sizes   : 38 bits physical, 48 bits virtual
power management:

processor       : 4
vendor_id       : GenuineIntel
cpu family      : 6
model           : 23
model name      : Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz
stepping        : 6
microcode       : 0x60f
cpu MHz         : 2000.000
cache size      : 6144 KB
physical id     : 0
siblings        : 4
core id         : 1
cpu cores       : 4
apicid          : 1
initial apicid  : 1
fpu             : yes
fpu_exception   : yes
cpuid level     : 10
wp              : yes
flags           : fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx lm constant_tsc arch_perfmon pebs bts rep_good nopl aperfmperf pni dtes64 monitor ds_cpl vmx est tm2 ssse3 cx16 xtpr pdcm dca sse4_1 lahf_lm tpr_shadow vnmi flexpriority dtherm
bogomips        : 5985.45
clflush size    : 64
cache_alignment : 64
address sizes   : 38 bits physical, 48 bits virtual
power management:

processor       : 5
vendor_id       : GenuineIntel
cpu family      : 6
model           : 23
model name      : Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz
stepping        : 6
microcode       : 0x60f
cpu MHz         : 2000.000
cache size      : 6144 KB
physical id     : 1
siblings        : 4
core id         : 1
cpu cores       : 4
apicid          : 5
initial apicid  : 5
fpu             : yes
fpu_exception   : yes
cpuid level     : 10
wp              : yes
flags           : fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx lm constant_tsc arch_perfmon pebs bts rep_good nopl aperfmperf pni dtes64 monitor ds_cpl vmx est tm2 ssse3 cx16 xtpr pdcm dca sse4_1 lahf_lm tpr_shadow vnmi flexpriority dtherm
bogomips        : 5985.04
clflush size    : 64
cache_alignment : 64
address sizes   : 38 bits physical, 48 bits virtual
power management:

processor       : 6
vendor_id       : GenuineIntel
cpu family      : 6
model           : 23
model name      : Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz
stepping        : 6
microcode       : 0x60f
cpu MHz         : 2000.000
cache size      : 6144 KB
physical id     : 0
siblings        : 4
core id         : 3
cpu cores       : 4
apicid          : 3
initial apicid  : 3
fpu             : yes
fpu_exception   : yes
cpuid level     : 10
wp              : yes
flags           : fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx lm constant_tsc arch_perfmon pebs bts rep_good nopl aperfmperf pni dtes64 monitor ds_cpl vmx est tm2 ssse3 cx16 xtpr pdcm dca sse4_1 lahf_lm tpr_shadow vnmi flexpriority dtherm
bogomips        : 5985.45
clflush size    : 64
cache_alignment : 64
address sizes   : 38 bits physical, 48 bits virtual
power management:

processor       : 7
vendor_id       : GenuineIntel
cpu family      : 6
model           : 23
model name      : Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz
stepping        : 6
microcode       : 0x60f
cpu MHz         : 2000.000
cache size      : 6144 KB
physical id     : 1
siblings        : 4
core id         : 3
cpu cores       : 4
apicid          : 7
initial apicid  : 7
fpu             : yes
fpu_exception   : yes
cpuid level     : 10
wp              : yes
flags           : fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx lm constant_tsc arch_perfmon pebs bts rep_good nopl aperfmperf pni dtes64 monitor ds_cpl vmx est tm2 ssse3 cx16 xtpr pdcm dca sse4_1 lahf_lm tpr_shadow vnmi flexpriority dtherm
bogomips        : 5985.04
clflush size    : 64
cache_alignment : 64
address sizes   : 38 bits physical, 48 bits virtual
power management:
`

	cpus = []*CpuInfo{
		&CpuInfo{"0", "GenuineIntel", "Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz", "2000.000", "6144 KB", "0", "0"},
		&CpuInfo{"1", "GenuineIntel", "Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz", "2000.000", "6144 KB", "1", "0"},
		&CpuInfo{"2", "GenuineIntel", "Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz", "2000.000", "6144 KB", "0", "2"},
		&CpuInfo{"3", "GenuineIntel", "Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz", "2000.000", "6144 KB", "1", "2"},
		&CpuInfo{"4", "GenuineIntel", "Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz", "2000.000", "6144 KB", "0", "1"},
		&CpuInfo{"5", "GenuineIntel", "Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz", "2000.000", "6144 KB", "1", "1"},
		&CpuInfo{"6", "GenuineIntel", "Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz", "2000.000", "6144 KB", "0", "3"},
		&CpuInfo{"7", "GenuineIntel", "Intel(R) Xeon(R) CPU           X5450  @ 3.00GHz", "2000.000", "6144 KB", "1", "3"},
	}
)

func fakeExecCommand(cmd string, args ...string) ([]byte, error) {
	return []byte(input), nil
}

func TestNewAllCpuInfo(t *testing.T) {
	execcmd = fakeExecCommand
	funcout, err := NewAllCpuInfo()
	if err != nil {
		t.Errorf("NewAllCpuInfo return error (%s)", err)
	}

	testutil.CompareVar(t, funcout, cpus)
}
