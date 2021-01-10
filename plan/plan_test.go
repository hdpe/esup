package plan

import (
	"context"
	"fmt"
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/es"
	"github.com/hdpe.me/esup/resource"
	"github.com/hdpe.me/esup/schema"
	"github.com/hdpe.me/esup/testutil"
	"github.com/hdpe.me/esup/util"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"testing"
)

func TestPlanner_Plan(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	testCases := indexSetTestCases

	c, err := NewElasticsearchContainer()

	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		if err := c.Terminate(); err != nil {
			println(err)
		}
	}()

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			s, err := tc.Schema()

			if err != nil {
				t.Error(err)
				return
			}

			defer func() {
				if err := tc.Clean(); err != nil {
					println(err)
				}
			}()

			plan, err := GetPlan(c, s, tc.Clock())

			if err != nil {
				t.Errorf("%w", err)
				return
			}

			if got, want := len(plan), len(tc.expected); got != want {
				t.Errorf("got %v action(s), want %v", got, want)
				return
			}

			for i, _ := range plan {
				if match := tc.Expected()[i].Match(plan[i]); !match.Matched {
					t.Errorf("%v", match.Failures)
				}
			}
		})
	}
}

func GetPlan(c *ElasticsearchContainer, s schema.Schema, clock util.Clock) ([]PlanAction, error) {
	baseUrl, err := c.BaseUrl()

	if err != nil {
		return nil, err
	}

	conf := config.Config{
		Server: config.ServerConfig{
			Address: baseUrl,
		},
		Changelog: config.ChangelogConfig{
			Index: "changelog",
		},
	}

	client, err := es.NewClient(conf.Server)

	if err != nil {
		return nil, err
	}

	changelog := resource.NewChangelog(conf.Changelog, client)
	proc := resource.NewPreprocessor(conf.Preprocess)

	p := NewPlanner(client, conf, changelog, s, proc, clock)

	return p.Plan()
}

type ElasticsearchContainer struct {
	c   testcontainers.Container
	ctx context.Context
}

func (r *ElasticsearchContainer) BaseUrl() (string, error) {
	host, err := r.c.Host(r.ctx)
	if err != nil {
		return "", err
	}

	port, err := r.c.MappedPort(r.ctx, "9200")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%s", host, port.Port()), nil
}

func (r *ElasticsearchContainer) Terminate() error {
	return r.c.Terminate(r.ctx)
}

func NewElasticsearchContainer() (*ElasticsearchContainer, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "elasticsearch:7.9.3",
		ExposedPorts: []string{"9200/tcp"},
		Env:          map[string]string{"discovery.type": "single-node"},
		WaitingFor: wait.ForHTTP("/_cluster/health").WithPort("9200/tcp").WithResponseMatcher(func(body io.Reader) bool {
			bytes, err := ioutil.ReadAll(body)

			if err != nil {
				return false
			}

			return gjson.GetBytes(bytes, "status").String() == "green"
		}),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	return &ElasticsearchContainer{
		c:   c,
		ctx: ctx,
	}, err
}

type planTestCase interface {
	desc() string
	envName() string
	schema() schema.Schema
	clock() util.Clock
	expected() []testutil.Matcher
}
