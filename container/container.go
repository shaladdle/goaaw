package container

type Pusher interface {
    Push(interface{})
}

type Popper interface {
    Pop() interface{}
}

type Queue interface {
    Pusher
    Popper
}
