package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"

	"github.com/oars-sigs/oars-cloud/core"
	resStore "github.com/oars-sigs/oars-cloud/pkg/store/resources"
)

type cronController struct {
	kv         core.KVStore
	cronStore  core.ResourceStore
	svcStore   core.ResourceStore
	cronLister core.ResourceLister
}

func newCronc(kv core.KVStore) *cronController {
	return &cronController{kv: kv}
}

func (c *cronController) run(stopCh chan struct{}) error {
	c.cronStore = resStore.NewStore(c.kv, new(core.Cron))
	c.svcStore = resStore.NewStore(c.kv, new(core.Service))
	cronLister, err := resStore.NewLister(c.kv, new(core.Cron), &core.ResourceEventHandle{})
	if err != nil {
		return err
	}
	c.cronLister = cronLister
	go c.start(stopCh)
	return nil
}

func (c *cronController) start(stopCh <-chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			resources, _ := c.cronLister.List()
			now := time.Now()
			ctx := context.TODO()
			for _, resource := range resources {
				job := resource.(*core.Cron)
				if job.Status == nil {
					job.Status = new(core.CronStatus)
				}
				if job.Disabled {
					continue
				}
				if now.Unix() < job.Status.Next {
					continue
				}
				sched, err := cron.Parse(job.Expr)
				if err != nil {
					logrus.Error(err)
					continue
				}

				job.Status.Prev = job.Status.Next
				job.Status.Next = sched.Next(now).Unix()
				_, err = c.cronStore.Put(ctx, job, &core.PutOptions{})
				if err != nil {
					logrus.Error(err)
					continue
				}
				if job.Service == nil {
					logrus.Error("service not defind")
					return
				}
				job.Service.ResourceMeta = &core.ResourceMeta{
					Name:      fmt.Sprintf("system-cron-%s", job.Name),
					Namespace: job.Namespace,
				}

				if job.Service.Docker.Labels == nil {
					job.Service.Docker.Labels = make(map[string]string)
				}
				job.Service.Docker.Labels[core.CronLabelKey] = strconv.Itoa(int(job.Status.Prev))
				_, err = c.svcStore.Put(ctx, job.Service, &core.PutOptions{})
				if err != nil {
					logrus.Error(err)
					continue
				}
				logrus.Info("finish cron ", job.Name)
			}
		}
	}
}
