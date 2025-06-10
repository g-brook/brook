package remote

// Plugin
// @Description: 一个插件的处理机制.
type Plugin interface {
	//
	// Name
	//  @Description: 描述插件的名字.
	//  @return string name.
	//
	Name() string
}
