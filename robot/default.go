package robot

import (
	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
)

type DefaultApi struct {
	znet.BaseRouter
}

func (*DefaultApi) Handle(request ziface.IRequest) {

}
