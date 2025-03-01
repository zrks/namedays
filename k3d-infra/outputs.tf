output "cluster_name" {
  description = "The name of the K3D cluster"
  value       = k3d_cluster.cluster.name
}

output "server_count" {
  description = "Number of server nodes in the cluster"
  value       = k3d_cluster.cluster.servers
}

output "agent_count" {
  description = "Number of agent nodes in the cluster"
  value       = k3d_cluster.cluster.agents
}

output "kube_api_host" {
  description = "The host of the Kubernetes API server"
  value       = k3d_cluster.cluster.kube_api[0].host
}

output "kube_api_host_ip" {
  description = "The IP address of the Kubernetes API host"
  value       = k3d_cluster.cluster.kube_api[0].host_ip
}

output "kube_api_host_port" {
  description = "The port of the Kubernetes API host"
  value       = k3d_cluster.cluster.kube_api[0].host_port
}

output "cluster_image" {
  description = "The image used for the K3D cluster"
  value       = k3d_cluster.cluster.image
}

output "cluster_network" {
  description = "The network name used by the cluster"
  value       = k3d_cluster.cluster.network
}

output "disable_load_balancer" {
  description = "Indicates if the load balancer is disabled"
  value       = k3d_cluster.cluster.k3d[0].disable_load_balancer
}

output "disable_host_ip_injection" {
  description = "Indicates if host IP injection is disabled"
  value       = k3d_cluster.cluster.k3d[0].disable_host_ip_injection
}

output "extra_server_args" {
  description = "Extra arguments passed to the K3S server"
  value       = k3d_cluster.cluster.k3s[0].extra_server_args
}
