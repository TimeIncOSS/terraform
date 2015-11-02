package terraform

import (
	"log"

	"github.com/hashicorp/terraform/config"
)

type EvalDebug struct {
	Event    string
	Resource *config.Resource
	One, Two **InstanceDiff
}

func (e EvalDebug) Eval(EvalContext) (interface{}, error) {
	if e.One == nil {
		log.Printf("[EvalDebug] %s: %s: Here is one: %#v", e.Resource.Id(), e.Event, e.One)
	} else {
		one := *e.One
		if one == nil {
			one = new(InstanceDiff)
			one.init()
		}

		log.Printf("[EvalDebug] %s: %s: Here is one: %#v", e.Resource.Id(), e.Event, one.Attributes["name"])
	}

	if e.Two == nil {
		log.Printf("[EvalDebug] %s: %s: Here is two: %#v", e.Resource.Id(), e.Event, e.Two)
	} else {
		two := *e.Two
		if two == nil {
			two = new(InstanceDiff)
			two.init()
		}

		log.Printf("[EvalDebug] %s: %s: Here is two: %#v", e.Resource.Id(), e.Event, two.Attributes["name"])
	}

	return nil, nil
}
