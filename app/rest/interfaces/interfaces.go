package interfaces

type JSON map[string]interface{}

type Api interface {
	Run() error
}
