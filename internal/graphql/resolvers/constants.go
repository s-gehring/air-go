package resolvers

// MaxBatchSize is the maximum number of identifiers allowed in a single byKeysGet request
// This limit protects system resources and ensures reasonable query performance
const MaxBatchSize = 100
