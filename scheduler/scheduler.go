/*
 *     Copyright 2020 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scheduler

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"

	managerv1 "d7y.io/api/pkg/apis/manager/v1"

	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/pkg/dfpath"
	"d7y.io/dragonfly/v2/pkg/gc"
	managerclient "d7y.io/dragonfly/v2/pkg/rpc/manager/client"
	"d7y.io/dragonfly/v2/scheduler/config"
	"d7y.io/dragonfly/v2/scheduler/job"
	"d7y.io/dragonfly/v2/scheduler/metrics"
	"d7y.io/dragonfly/v2/scheduler/resource"
	"d7y.io/dragonfly/v2/scheduler/rpcserver"
	"d7y.io/dragonfly/v2/scheduler/scheduler"
	"d7y.io/dragonfly/v2/scheduler/service"
	"d7y.io/dragonfly/v2/scheduler/storage"
)

const (
	// gracefulStopTimeout specifies a time limit for
	// grpc server to complete a graceful shutdown.
	gracefulStopTimeout = 10 * time.Minute
)

type Server struct {
	// Server configuration.
	config *config.Config

	// GRPC server.
	grpcServer *grpc.Server

	// Metrics server.
	metricsServer *http.Server

	// Manager client.
	managerClient managerclient.Client

	// Dynamic config.
	dynconfig config.DynconfigInterface

	// Async job.
	job job.Job

	// Storage service.
	storage storage.Storage

	// GC server.
	gc gc.GC
}

func New(ctx context.Context, cfg *config.Config, d dfpath.Dfpath) (*Server, error) {
	s := &Server{config: cfg}

	// Initialize manager client.
	managerClient, err := managerclient.GetClient(ctx, cfg.Manager.Addr)
	if err != nil {
		return nil, err
	}
	s.managerClient = managerClient

	// Register to manager.
	if _, err := s.managerClient.UpdateScheduler(context.Background(), &managerv1.UpdateSchedulerRequest{
		SourceType:         managerv1.SourceType_SCHEDULER_SOURCE,
		HostName:           s.config.Server.Host,
		Ip:                 s.config.Server.IP,
		Port:               int32(s.config.Server.Port),
		Idc:                s.config.Host.IDC,
		Location:           s.config.Host.Location,
		SchedulerClusterId: uint64(s.config.Manager.SchedulerClusterID),
	}); err != nil {
		logger.Fatalf("register to manager failed %s", err.Error())
	}

	// Initialize dynconfig client.
	dynconfig, err := config.NewDynconfig(s.managerClient, d.CacheDir(), cfg)
	if err != nil {
		return nil, err
	}
	s.dynconfig = dynconfig

	// Initialize GC.
	s.gc = gc.New(gc.WithLogger(logger.GCLogger))

	// Initialize resource.
	resource, err := resource.New(cfg, s.gc, dynconfig)
	if err != nil {
		return nil, err
	}

	// Initialize scheduler.
	scheduler := scheduler.New(cfg.Scheduler, dynconfig, d.PluginDir())

	// Initialize Storage.
	storage, err := storage.New(d.DataDir())
	if err != nil {
		return nil, err
	}
	s.storage = storage

	// Initialize scheduler service.
	service := service.New(cfg, resource, scheduler, dynconfig, s.storage)

	// Initialize grpc service.
	svr := rpcserver.New(service)
	s.grpcServer = svr

	// Initialize job service.
	if cfg.Job.Enable {
		s.job, err = job.New(cfg, resource)
		if err != nil {
			return nil, err
		}
	}

	// Initialize metrics.
	if cfg.Metrics.Enable {
		s.metricsServer = metrics.New(cfg.Metrics, s.grpcServer)
	}

	return s, nil
}

func (s *Server) Serve() error {
	// Serve dynConfig.
	go func() {
		if err := s.dynconfig.Serve(); err != nil {
			logger.Fatalf("dynconfig start failed %s", err.Error())
		}
		logger.Info("dynconfig start successfully")
	}()

	// Serve GC.
	s.gc.Serve()
	logger.Info("gc start successfully")

	// Serve Job.
	if s.config.Job.Enable {
		s.job.Serve()
		logger.Info("job start successfully")
	}

	// Started metrics server.
	if s.metricsServer != nil {
		go func() {
			logger.Infof("started metrics server at %s", s.metricsServer.Addr)
			if err := s.metricsServer.ListenAndServe(); err != nil {
				if err == http.ErrServerClosed {
					return
				}
				logger.Fatalf("metrics server closed unexpect: %s", err.Error())
			}
		}()
	}

	if s.managerClient != nil {
		// scheduler keepalive with manager.
		go func() {
			logger.Info("start keepalive to manager")
			s.managerClient.KeepAlive(s.config.Manager.KeepAlive.Interval, &managerv1.KeepAliveRequest{
				SourceType: managerv1.SourceType_SCHEDULER_SOURCE,
				HostName:   s.config.Server.Host,
				Ip:         s.config.Server.IP,
				ClusterId:  uint64(s.config.Manager.SchedulerClusterID),
			})
		}()
	}

	// Generate GRPC limit listener.
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.Server.Listen, s.config.Server.Port))
	if err != nil {
		logger.Fatalf("net listener failed to start: %s", err.Error())
	}
	defer listener.Close()

	// Started GRPC server.
	logger.Infof("started grpc server at %s://%s", listener.Addr().Network(), listener.Addr().String())
	if err := s.grpcServer.Serve(listener); err != nil {
		logger.Errorf("stoped grpc server: %s", err.Error())
		return err
	}

	return nil
}

func (s *Server) Stop() {
	// Stop dynconfig server.
	if err := s.dynconfig.Stop(); err != nil {
		logger.Errorf("dynconfig client closed failed %s", err.Error())
	} else {
		logger.Info("dynconfig client closed")
	}

	// Clean storage.
	if err := s.storage.Clear(); err != nil {
		logger.Errorf("clean storage failed %s", err.Error())
	} else {
		logger.Info("clean storage completed")
	}

	// Stop GC.
	s.gc.Stop()
	logger.Info("gc closed")

	// Stop metrics server.
	if s.metricsServer != nil {
		if err := s.metricsServer.Shutdown(context.Background()); err != nil {
			logger.Errorf("metrics server failed to stop: %s", err.Error())
		} else {
			logger.Info("metrics server closed under request")
		}
	}

	// Stop GRPC server.
	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		logger.Info("grpc server closed under request")
		close(stopped)
	}()

	t := time.NewTimer(gracefulStopTimeout)
	select {
	case <-t.C:
		s.grpcServer.Stop()
	case <-stopped:
		t.Stop()
	}
}
