package sphinx

import "github.com/google/uuid"

// BranchGroupContent represents a type aliasing a map
// where key is a branch group UUID and value is a slice of UUIDs.
// Each of latter UUIDs is a ID of a branch belonging to the branch group
// designated by the key. The value contains all the branches of the branch group.
type BranchGroupContent map[uuid.UUID][]uuid.UUID

// BranchGroupsOfBranch represents a type aliasing a map
// where key is a branch UUID and value is a slice of UUIDs.
// Each of latter UUIDs is an ID of a branch group containing the branch
// designated by the key. The value contains all the branch groups the branch belongs to.
type BranchGroupsOfBranch map[uuid.UUID][]uuid.UUID

func (m BranchGroupContent) Reverse() BranchGroupsOfBranch {
	result := make(BranchGroupsOfBranch)
	for group, branches := range m {
		for _, branch := range branches {
			result[branch] = append(result[branch], group)
		}
	}
	return result
}
