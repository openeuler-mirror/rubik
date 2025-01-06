
# QuotaTurbo API Documentation

我们提供了quotaTubro底层库（`/pkg/lib/cpu/quotatubro`），用户可以通过vendor第三方库等方式源码引用，实现弹性限流功能。

quotaTurbo对外提供如下稳定API：
- **实例创建**
  - [func NewClient() ClientAPI](#func-newclient)
- **配额调整**
  - [AdjustQuota() error](#func-adjustquota)
- **Cgroup生命周期管理**
  - [AddCgroup(string, float64) error](#func-addcgroup)
  - [RemoveCgroup(string) error](#func-removecgroup)
  - [AllCgroups() []string](#func-allcgroups)
- **参数管理**
  - 设置
    - [WithOptions(...Option) error](#func-withoptions)
    - [WithWaterMark(highVal, alarmVal int) Option](#func-withwatermark)
    - [WithCgroupRoot(string) Option](#func-withcgrouproot)
    - [WithElevateLimit(float64) Option](#func-withelevatelimit)
    - [WithSlowFallbackRatio(float64) Option](#func-withslowfallbackratio)
    - [WithCPUFloatingLimit(float64) Option](#func-withcpufloatinglimit)
  - 查询
    - [AlarmWaterMark() int](#func-alarmwatermark)
    - [HighWaterMark() int](#func-highwatermark)
    - [CgroupRoot() string](#func-cgrouproot)
    - [ElevateLimit() float64](#func-elevatelimit)
    - [SlowFallbackRatio() float64](#func-slowfallbackratio)
    - [CPUFloatingLimit() float64](#func-cpufloatinglimit)

## Types

### type ConfigViewer
```golang
type ConfigViewer interface {
	AlarmWaterMark() int
	HighWaterMark() int
	CgroupRoot() string
	ElevateLimit() float64
	SlowFallbackRatio() float64
	CPUFloatingLimit() float64
}
```
`ConfigViewer`接口用于规范quotaTurbo参数查询方法。

### type ClientAPI
```golang
type ClientAPI interface {
	// Quota detection and adjustment
	AdjustQuota() error
	// Cgroup lifecycle management
	AddCgroup(string, float64) error
	RemoveCgroup(string) error
	AllCgroups() []string
	// Parameter management
	WithOptions(...Option) error
	ConfigViewer
}
```
`ClientAPI`接口用于规范quotaTurbo的配额调整、Cgroup生命周期管理、参数管理方法。

### type Config
```golang
type Config struct {
	HighWaterMark int
	AlarmWaterMark int
	CgroupRoot string
	ElevateLimit float64
	SlowFallbackRatio float64
	CPUFloatingLimit float64
}
```
`Config`是quotaTurbo参数。

### type Option
```golang
type Option func(c *Config) error
```
`Option`是设置quotaTurbo参数的函数接口。

### type Client 
```golang
struct {
	// 其成员变量大写，但目前并未开放给用户自由使用，
}
```
我们提供了`NewClient`方法供外部客户新建Client，其内部参数约束为不可直接使用，不建议用户直接调用。

#### *func NewClient*
```golang
func NewClient() ClientAPI
```
NewClient用于创建一个ClientAPI实例，**建议仅使用该方法获取client实例**。

#### *func AdjustQuota*

```golang
func (c *Client)AdjustQuota() error
```
AdjustQuota根据当前算法参数以及当前cgroup的运行状态调整每一个cgroup最大可用cpu时间片（即`cpu.cfs_quota_period`）。

#### *func AddCgroup*

```golang
func (c *Client)AddCgroup(string, float64) error
```
AddCgroup为quotaTurbo添加目标cgroup。第一个参数是cgroup相对于cgroup挂载点的路径。例如，容器的cgroup绝对路径为`/sys/fs/cgroup/kubepods/podxxx/yyy`，则输入参数为`kubepods/podxxx/yyy`。第二个参数为指定cgroup的cpu限制，其含义即取值同kubernetes的[cpulimit](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#requests-and-limits)。

> 注：若cpulimit为0或者超过节点最大可用核数，则指定cgroup将不会被调整。

#### *func RemoveCgroup*
```golang
func (c *Client)RemoveCgroup(string) error
```
RemoveCgroup删除指定cgroup，不再对该cgroup进行配额调整。接收参数为指定cgroup相对于cgroup挂载目录的路径，同[AddCgroup](#func-addcgroup)。

#### *func AllCgroups*
```golang
func (c *Client)AllCgroups() []string
```
AllCgroups返回当前正在调整的cgroup的相对路径列表。

#### *func AlarmWaterMark*
```golang
func (c *Client)AlarmWaterMark() int
```
AlarmWaterMark返回当前quotaTurbo警戒水位参数取值。

#### *func HighWaterMark*
```golang
func (c *Client)HighWaterMark() int
```
HighWaterMark返回当前quotaTurbo高水位参数取值。

#### *func CgroupRoot*
```golang
func (c *Client)CgroupRoot() string
```
CgroupRoot返回当前quotaTurbo cgroup挂载点参数取值。

#### *func ElevateLimit*
```golang
func (c *Client)ElevateLimit() float64
```
ElevateLimit返回当前quotaTurbo单次缓提升上限参数取值。

#### *func SlowFallbackRatio*
```golang
func (c *Client)SlowFallbackRatio() float64
```
SlowFallbackRatio返回当前quotaTurbo单次慢回调比率参数取值。

#### *func CPUFloatingLimit*
```golang
func (c *Client)CPUFloatingLimit() float64
```
CPUFloatingLimit返回当前quotaTurbo一分钟内CPU最大变化幅度参数取值。

#### *func WithOptions*
```golang
func (c *Client)WithOptions(...Option) error
```
WithOptions接收参数为Option列表，Option定义为`type Option func(c *Config) error`，用于设置参数。quotaTubro提供了默认的Option，用法如下：
```golang
c := NewClient()
// 设置cgroup挂载点
c.WithOptions(WithCgroupRoot(constant.TmpTestDir))
```


## Functions
### *func WithWaterMark*
```golang
func WithWaterMark(highVal, alarmVal int) Option
```
WithWaterMark提供了设置水位线参数的Option。第一个参数是高水位，第二个参数是警戒水位，二者协同控制缓提升、慢回调、快回落算法，调整cgroup最大可用cpu时间片。其中，参数需满足 $0<=highVal<alarmVal<=100$。高水位默认值为60，警戒水位默认为80。

### *func WithCgroupRoot*
```golang
func WithCgroupRoot(string) Option Option
```
WithCgroupRoot提供了设置Cgroup挂载点参数`CgroupRoot`的Option。目前rubik仅支持cgroupfs，默认值为`/sys/fs/cgroup`。

### *func WithElevateLimit*
```golang
func WithElevateLimit(float64) Option
```
WithElevateLimit提供了设置单次缓提升总CPU提升上限参数`ElevateLimit`的Option。该参数用于控制单次缓提升的最大百分比，取值范围为$[0,100]$，默认值为1。

### *func WithSlowFallbackRatio*
```golang
func WithSlowFallbackRatio(float64) Option
```
WithSlowFallbackRatio提供了设置单次慢回调下降比率参数`SlowFallbackRatio`的Option。该参数越大，单次慢回调下降越快。默认值为0.1。

### *func WithCPUFloatingLimit*
```golang
func WithCPUFloatingLimit(float64) Option
```
WithCPUFloatingLimit提供了设置允许CPU最大变化幅度参数`CPUFloatingLimit`的Option。quotaTurbo会统计1min内整机节点CPU利用率变动幅度，若CPU变化幅度超过该参数，则不进行缓提升。`CPUFloatingLimit`取值范围为$[0,100]$，默认取值为10。