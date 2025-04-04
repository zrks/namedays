terraform {
  required_providers {
    k3d = {
      source = "pvotal-tech/k3d"
      version = "0.0.5"
    }
  }
}

resource "k3d_cluster" "cluster" {
  name    = "yanka" # var.cluster_name
  servers = 1
  agents  = 3 # var.node_count

  kube_api {
    host      = "k8s.test.localhost" # "k8s.${var.cluster_env}.localhost"
    host_ip   = "0.0.0.0"
    host_port = "6443" # var.cluster_api_port
  }

  image   = "rancher/k3s:v1.24.17-k3s1" # "rancher/k3s:${var.cluster_version}"
  network = "test-k3d" # "${var.cluster_env}-k3d"

  registries {}

  k3d {
    disable_load_balancer = true
    disable_host_ip_injection = false
  }

  k3s {
    extra_server_args = [
      "--tls-san=k8s.test.localhost", # "--tls-san=k8s.${var.cluster_env}.localhost",
      "--tls-san=host.k3d.internal",
      "--disable=servicelb",
      "--disable=traefik",
    ]
  }
}
