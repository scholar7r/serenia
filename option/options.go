package option

import (
	"context"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/json/badoption"
)

type Profile struct {
	Name      string                     `json:"name,omitempty"`
	Outbounds badoption.Listable[string] `json:"outbounds,omitempty"`
}

type User struct {
	Name     string                     `json:"name,omitempty"`
	UUID     string                     `json:"uuid,omitempty"`
	Profiles badoption.Listable[string] `json:"profiles,omitempty"`
}

func (u User) Variables() map[string]string {
	return map[string]string{
		"UUID": u.UUID,
	}
}

type _Options struct {
	RawMessage json.RawMessage                       `json:"-"`
	Log        *option.LogOptions                    `json:"log,omitempty"`
	Listen     string                                `json:"listen,omitempty"`
	Outbounds  []badoption.Listable[option.Outbound] `json:"outbounds,omitempty"`
	Profiles   []Profile                             `json:"profiles,omitempty"`
	Users      []User                                `json:"users,omitempty"`
}

type Options _Options

func (o *Options) UnmarshalJSONContext(ctx context.Context, content []byte) error {
	if err := json.UnmarshalContextDisallowUnknownFields(ctx, content, (*_Options)(o)); err != nil {
		return err
	}
	o.RawMessage = content
	return nil
}

var _ json.ContextUnmarshaler = (*Options)(nil)
