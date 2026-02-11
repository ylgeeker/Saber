//go:build integration

/**
 * Copyright 2025 Saber authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

package discovery_test

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"os-artificer/saber/pkg/discovery"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var client *discovery.Client
var etcdClient *clientv3.Client
var reg *discovery.Registry
var dis *discovery.Discovery

func setup() {

	endpoints := os.Getenv("SABER_ETCD_ENDPOINTS")
	user := os.Getenv("SABER_ETCD_USER")
	password := os.Getenv("SABER_ETCD_PASSWORD")

	log.Println("endpoints:", endpoints)
	log.Println("user:", user)
	log.Println("password:", password)

	if endpoints == "" {
		log.Fatal("endpoints is required")
	}

	opts := []discovery.Option{
		discovery.OptionEndpoints(strings.Split(endpoints, ";")),
	}

	// Support no-auth mode
	if user != "" && password != "" {
		opts = append(opts, discovery.OptionUser(user), discovery.OptionPassword(password))
	}

	cli, err := discovery.NewClientWithOptions(opts...)

	if err != nil {
		log.Fatalf("failed to create etcd client. errmsg:%s", err.Error())
	}

	client = cli
	etcdClient, err = cli.OriginClient()
	if err != nil {
		log.Fatalf("failed to create etcd client, errmsg: %v", err)
	}

	reg = cli.CreateRegistry()

	d, err := cli.CreateDiscovery()
	if err != nil {
		log.Fatalf("failed to create discovery. errmsg:%s", err.Error())
	}

	err = reg.SetService(context.Background(), "")
	if err != nil {
		log.Fatalf("failed to set service. errmsg:%s", err.Error())
	}

	dis = d
}

func teardown() {
	reg.Close()
	dis.Close()
}

func TestRegistrySetService(t *testing.T) {

	err := reg.SetService(context.Background(), "test-id")
	if err != nil {
		t.Logf("failed to set service. errmsg:%s", err.Error())
	}
}

func TestDiscotryGetWithPrefix(t *testing.T) {

	kvs, err := dis.GetWithPrefix(context.Background(), "")
	if err != nil {
		t.Logf("failed to get service. errmsg:%s", err.Error())
	}

	for key, value := range kvs {
		t.Logf("key:%s, value:%s", key, value)
	}
}

func TestMain(m *testing.M) {

	setup()

	code := m.Run()

	teardown()

	os.Exit(code)
}
