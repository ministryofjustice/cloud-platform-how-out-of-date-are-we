# Cloud Platform Cost Calculator

Use the [AWS CostExplorer][cost explorer] to work out the running costs (in US
Dollars) of namespaces in the Cloud Platform.

Costs are allocated based on the value of the `namespace` tag.

Monthly costs are calculated by multiplying yesterday's cost data by 30. This
means you don't have to wait for a month to get usable monthly costs, but it
will exacerbate the apparent impact of temporary changes.

## Shared Support Costs

A portion of the total cost of the Cloud Platform is allocated to each namespace.
The total monthly team cost is hard-coded into the script and a portion (total
/ number of namespaces) is added to each namespace.
