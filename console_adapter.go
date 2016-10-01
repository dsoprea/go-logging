package log

import (
    golog "log"
)

type ConsoleLogAdapter struct {

}

func (cla *ConsoleLogAdapter) Debugf(lc *LogContext, message *string) error {
    golog.Println(*message)
//    a.Debugf(lc.Ctx, *message)

    return nil
}

func (cla *ConsoleLogAdapter) Infof(lc *LogContext, message *string) error {
    golog.Println(*message)
//    a.Infof(lc.Ctx, *message)

    return nil
}

func (cla *ConsoleLogAdapter) Warningf(lc *LogContext, message *string) error {
    golog.Println(*message)
//    a.Warningf(lc.Ctx, *message)

    return nil
}

func (cla *ConsoleLogAdapter) Errorf(lc *LogContext, message *string) error {
    golog.Println(*message)
//    a.Errorf(lc.Ctx, *message)

    return nil
}


type ConsoleAdapterMaker struct {

}

func NewConsoleAdapterMaker() *ConsoleAdapterMaker {
    return new(ConsoleAdapterMaker)
}

func (cam ConsoleAdapterMaker) New() LogAdapter {
    return new(ConsoleLogAdapter)
}
