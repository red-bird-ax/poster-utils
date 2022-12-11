package utils

type RepoRequestOptions struct {
    Pagination struct{
        Offset int64
        Amount int64
    }
    OrderBy string
    Asc     bool
}