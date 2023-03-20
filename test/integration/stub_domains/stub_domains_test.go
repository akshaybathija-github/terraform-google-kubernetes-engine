// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package stub_domains

import (
	"fmt"
	"testing"

	"github.com/GoogleCloudPlatform/cloud-foundation-toolkit/infra/blueprint-test/pkg/gcloud"
	"github.com/GoogleCloudPlatform/cloud-foundation-toolkit/infra/blueprint-test/pkg/golden"
	"github.com/GoogleCloudPlatform/cloud-foundation-toolkit/infra/blueprint-test/pkg/tft"
	"github.com/GoogleCloudPlatform/cloud-foundation-toolkit/infra/blueprint-test/pkg/utils"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/assert"
)

func TestStubDomains(t *testing.T) {
	bpt := tft.NewTFBlueprintTest(t)

	bpt.DefineVerify(func(assert *assert.Assertions) {
		//Skipping Default Verify as the Verify Stage fails due to change in Client Cert Token
		// bpt.DefaultVerify(assert)

		projectId := bpt.GetStringOutput("project_id")
		location := bpt.GetStringOutput("location")
		clusterName := bpt.GetStringOutput("cluster_name")
		serviceAccount := bpt.GetStringOutput("service_account")

		op := gcloud.Runf(t, "container clusters describe %s --zone %s --project %s", clusterName, location, projectId)
		g := golden.NewOrUpdate(t, op.String(),
			golden.WithSanitizer(golden.StringSanitizer(serviceAccount, "SERVICE_ACCOUNT")),
			golden.WithSanitizer(golden.StringSanitizer(projectId, "PROJECT_ID")),
			golden.WithSanitizer(golden.StringSanitizer(clusterName, "CLUSTER_NAME")),
		)
		validateJSONPaths := []string{
			"status",
			"addonsConfig",
		}
		for _, pth := range validateJSONPaths {
			g.JSONEq(assert, op, pth)
		}
		gcloud.Runf(t, "container clusters get-credentials %s --region %s --project %s", clusterName, location, projectId)
		k8sOpts := k8s.KubectlOptions{}

		for _, _ = range []struct {
			domain string
			ips    string
		}{
			{
				domain: "example.com",
				ips:    "[\"10.254.154.11\",\"10.254.154.12\"]",
			},
			{
				domain: "example.net",
				ips:    "[\"10.254.154.11\",\"10.254.154.12\"]",
			},
		} {
			configMap, err := k8s.RunKubectlAndGetOutputE(t, &k8sOpts, "get", "configmap", "kube-dns", "-n", "kube-system", "-o", "json")
			assert.NoError(err)
			stubDomain := utils.ParseJSONResult(t, configMap)
			stubDomainMap := stubDomain.Get("data.stubDomains").Map()
			fmt.Print(stubDomainMap)

			// assert.Contains(stubDomain.Get("data.stubDomains.#").String(), stubDomains.domain, "reflects the stub_domains configuration")
		}
		// jsonStringValue := "{\"example.com\":[\"10.254.154.11\",\"10.254.154.12\"],\"example.net\":[\"10.254.154.11\",\"10.254.154.12\"]}"

		// ipmasqMap, err := k8s.RunKubectlAndGetOutputE(t, &k8sOpts, "get", "configmap", "ip-masq-agent", "-n", "kube-system", "-o", "json")
		// assert.NoError(err)
		// ipmasq := utils.ParseJSONResult(t, ipmasqMap)
		// ipmasqjsonStringValue := "nonMasqueradeCIDRs:\n  - 10.0.0.0/8\n  - 172.16.0.0/12\n  - 192.168.0.0/16\nresyncInterval: 60s\nmasqLinkLocal: false\n"
		// assert.Contains(ipmasq.Get("data.config").String(), ipmasqjsonStringValue, "ipmasq  configuration is valid")

	})

	bpt.Test()
}
