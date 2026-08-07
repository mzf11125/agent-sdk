package erc8004

// MetadataEntry is one on-chain metadata key/value pair stored with an agent
// registration (ERC-8004 IIdentityRegistry.MetadataEntry).
type MetadataEntry struct {
	MetadataKey   string
	MetadataValue []byte
}
