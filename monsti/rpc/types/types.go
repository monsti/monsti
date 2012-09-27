// Structs used for RPC calls.
package types

type GetNodeDataArgs struct {
  Path, File string
}

type WriteNodeDataArgs struct {
  Path, File, Content string
}
