# Orphaned AWS Resources

List all the AWS resources which exist but are not mentioned in any of the
`terraform.tfstate` files in the Cloud Platform terraform state S3 bucket.

1. get all the terraform.tfstate files

1. create arrays for:
    - kops clusters
    - vpcs
    - nat_gateways
    - subnets
    - route_tables
    - route_table_associations
    - internet_gateways
    - hosted_zones
    - rds
    - rds_cluster
    - NS records are orphaned in the cloud-platform.service.justice.gov.uk hosted zone (new)

1. get existing aws resources for and verify if they exist in the tfstate arrays:
    - kops_clusters
    - vpcs
    - nat_gateways
    - subnets
    - route_tables
    - route_table_associations
    - internet_gateways
    - hosted_zones
    - rds
    - rds_cluster
    - NS records are orphaned in the cloud-platform.service.justice.gov.uk hosted zone (new)

1. write the results to json and then push to an s3 bucket:

    ```
    {
        orphaned_aws_resources: {
            vpcs: compare(:vpcs),
            nat_gateways: compare(:nat_gateways),
            hosted_zones: compare(:hosted_zones),
            internet_gateways: compare(:internet_gateways),
            subnets: compare(:subnets),
            route_tables: compare(:route_tables),
            route_table_associations: compare(:route_table_associations),
            rds: compare(:rds),
            rds_cluster: compare(:rds_cluster),
            kops_cluster: orphaned_kops_clusters,
            ns_records: orphaned_aws_ns_records,
        },
        updated_at: Time.now()
    }
    ```
    
