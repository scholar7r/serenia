package profile

import (
	"bytes"
	"context"
	"text/template"

	boxOption "github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/json/badjson"
	"github.com/sagernet/sing/common/logger"
	"github.com/scholar7r/serenia/option"
)

type Profile struct {
	option.Profile
	manager *ProfileManager
}

func (p *Profile) Render(vars map[string]string) (*boxOption.Options, error) {
	outbounds := common.FlatMap(p.manager.outbounds, func(it []boxOption.Outbound) []boxOption.Outbound {
		if !common.Contains(p.Outbound, it[0].Tag) {
			return nil
		}
		return it
	})

	options := boxOption.Options{
		Outbounds: outbounds,
	}

	content, err := json.MarshalContext(p.manager.ctx, options)
	if err != nil {
		return nil, E.Cause(err, "marshal template")
	}

	tmpl, err := template.New(p.Name).Option("missingkey=error").Parse(string(content))
	if err != nil {
		return nil, E.Cause(err, "parse template")
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, vars); err != nil {
		return nil, E.Cause(err, "render template")
	}

	options, err = json.UnmarshalExtendedContext[boxOption.Options](p.manager.ctx, rendered.Bytes())
	if err != nil {
		return nil, E.Cause(err, "unmarshal rendered template")
	}

	options, err = badjson.Omitempty(p.manager.ctx, options)
	if err != nil {
		return nil, E.Cause(err, "omitempty")
	}

	return &options, nil
}

type ProfileManager struct {
	ctx       context.Context
	logger    logger.Logger
	outbounds [][]boxOption.Outbound
	profiles  []*Profile
}

func NewProfileManager(
	ctx context.Context,
	logger logger.Logger,
	outbounds [][]boxOption.Outbound,
	profiles []option.Profile,
) (*ProfileManager, error) {
	manager := &ProfileManager{
		ctx:       ctx,
		logger:    logger,
		outbounds: outbounds,
	}

	for i, v := range profiles {
		if v.Name == "" {
			return nil, E.New("initialize profile [", i, "]: missing name")
		}

		manager.profiles = append(manager.profiles, &Profile{
			Profile: v,
			manager: manager,
		})
	}

	return manager, nil
}

func (m *ProfileManager) ProfileByName(name string) *Profile {
	for _, v := range m.profiles {
		if v.Name == name {
			return v
		}
	}

	return nil
}
