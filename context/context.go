package context

type IContext interface {
	IContextimpl()
}

type Context struct{}
