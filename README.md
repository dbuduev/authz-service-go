## DynamoDB structure

Legend:

* HK: Hash Key
* RK: Range Key
* GSI: Global Secondary Index
* LSI: Local Secondary Index

| globalId (HK)              | applicationId (GSK1: HK) | id      | typeTarget (RK, GSK1: RK)                | typeTargetTagless  (LSK1: RK)            | data           |
| -------------------------- | -------------------------|-------- |------------------------------------------|------------------------------------------|----------------|
| '0_op1'                    | '0'                      | 'op1'   | 'node_OP &#124; op1'                           | 'node_OP &#124; op1'                           | 'add-member'   |
| '0_op1'                    | '0'                      | 'op1'   | 'edge_ROLE &#124; r1'                          | 'edge_ROLE &#124; r1'                          | 'r1'           |
| '0_r1'                     | '0'                      | 'r1'    | 'node_ROLE &#124; r1'                          | 'node_ROLE &#124; r1'                          | 'staff-member' |
| '0_r1'                     | '0'                      | 'r1'    | 'edge_OP &#124; op1'                           | 'edge_OP &#124; op1'                           | 'op1'          |
| '0_b1'                     | '0'                      | 'b1'    | 'node_BRANCH &#124; b1'                        | 'node_BRANCH &#124; b1'                        | undefined      |
| '0_g1'                     | '0'                      | 'g1'    | 'node_BRANCH_GROUP &#124; g1'                  | 'node_BRANCH_GROUP &#124; g1'                  | undefined      |
| '0_b1'                     | '0'                      | 'b1'    | 'edge_BRANCH_GROUP &#124; g1'                  | 'edge_BRANCH_GROUP &#124; g1'                  | 'g1'           |
| '0_g1'                     | '0'                      | 'g1'    | 'edge_BRANCH &#124; b1'                        | 'edge_BRANCH &#124; b1'                        | 'b1'           |
| '0_u1'                     | '0'                      | 'u1'    | 'edge_ROLE &#124; r1 &#124; ASSIGNED_IN_BRANCH &#124; b1'  | 'edge_ROLE &#124; r1'                          | 'b1'           |
| '0_r1'                     | '0'                      | 'r1'    | 'edge_USER &#124; u1 &#124; ASSIGNED_IN_BRANCH &#124; b1'  | 'edge_USER &#124; u1'                          | 'b1'           |

### Usage
`make test` will try to run Amazon DynamoDB container locally before running tests.
Requires Docker and AWS CLI.