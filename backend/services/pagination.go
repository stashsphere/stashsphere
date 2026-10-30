package services

// totalPageCount is the number of pages for count results. Without
// pagination all results are on a single page, whatever perPage says.
func totalPageCount(count uint64, perPage uint64, paginate bool) uint64 {
	if !paginate || perPage == 0 {
		return min(count, 1)
	}
	return (count + perPage - 1) / perPage
}
