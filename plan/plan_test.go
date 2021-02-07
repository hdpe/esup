package plan

import (
	"context"
	"fmt"
	"github.com/hdpe.me/esup/config"
	esupContext "github.com/hdpe.me/esup/context"
	"github.com/hdpe.me/esup/es"
	"github.com/hdpe.me/esup/resource"
	"github.com/hdpe.me/esup/schema"
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

	var testCases []PlanTestCase
	for _, tc := range indexSetTestCases {
		testCases = append(testCases, tc)
	}
	for _, tc := range documentTestCases {
		testCases = append(testCases, tc)
	}

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
		t.Run(tc.Desc(), func(t *testing.T) {
			s, err := tc.Schema()

			if err != nil {
				t.Fatal(err)
			}

			defer func() {
				if err := tc.Clean(); err != nil {
					println(err)
				}
			}()

			ctx, err := GetContext(c, s)

			if err != nil {
				t.Fatalf("%v", err)
			}

			if tc.Setup() != nil {
				coll := NewCollector()
				defer CleanUp(ctx, coll)

				tc.Setup()(Setup{
					es:        ctx.Es,
					changelog: ctx.Changelog,
					onError: func(err error) {
						t.Fatalf("error in test setup: %v", err)
					},
					collector: coll,
				})
			}

			p := NewPlanner(ctx.Es, ctx.Conf, ctx.Changelog, s, ctx.Proc, tc.Version())

			plan, err := p.Plan()

			if err != nil {
				t.Fatalf("%v", err)
			}

			if got, want := len(plan), len(tc.Expected()); got != want {
				t.Fatalf("got %v action(s), want %v", got, want)
			}

			for i, _ := range plan {
				if match := tc.Expected()[i].Match(plan[i]); !match.Matched {
					t.Errorf("%v", match.Failures)
				}
			}
		})
	}
}

func CleanUp(ctx *esupContext.Context, coll *Collector) {
	logOnError := func(f func() error) {
		if err := f(); err != nil {
			print(fmt.Errorf("%w", err))
		}
	}

	for _, p := range coll.Pipelines {
		logOnError(func() error { return ctx.Es.DeletePipeline(p) })
	}
	for _, i := range coll.Indices {
		logOnError(func() error { return ctx.Es.DeleteIndex(i) })
	}
	logOnError(func() error { return ctx.Es.DeleteIndex(ctx.Conf.Changelog.Index) })
}

func GetContext(c *ElasticsearchContainer, s schema.Schema) (*esupContext.Context, error) {
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

	return &esupContext.Context{
		Conf:      conf,
		Schema:    s,
		Es:        client,
		Changelog: changelog,
		Proc:      proc,
	}, nil
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
