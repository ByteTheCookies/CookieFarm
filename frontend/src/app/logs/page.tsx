'use client';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { columns } from './column';
import { DataTable } from './data-table';
import { FilterPanel } from './filter-panel';
import { usePaginatedFlags } from '@/hooks/usePaginatedFlags';

export default function FlagLogs() {
  const {
    data: filteredData,
    totalCount,
    error,
    pagination,
    filters,
    sorting,
    columnFilters,

    setPagination,
    setSearchTerm,
    setSearchField,
    setStatusFilter,
    setServiceFilter,
    setTeamFilter,
    setSortField,
    setSortDirection,
    setSorting,
    setColumnFilters,

    onClearAll,
  } = usePaginatedFlags();

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="flex flex-col items-center space-y-4">
          <div className="text-lg font-semibold text-red-500">
            Error loading flags
          </div>
          <p className="text-muted-foreground text-sm">
            {error.message || 'Failed to fetch flag data. Please try again.'}
          </p>
          <button
            onClick={() => window.location.reload()}
            className="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto space-y-6 p-4">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
            Flag Logs
          </h1>
          <p className="text-muted-foreground text-sm sm:text-base">
            Advanced server-side filtering, searching, sorting and pagination
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="outline" className="text-xs sm:text-sm">
            {filteredData.length} results
          </Badge>
          <Badge variant="secondary" className="text-xs sm:text-sm">
            Page {pagination.pageIndex + 1} of{' '}
            {Math.ceil(totalCount / pagination.pageSize)}
          </Badge>
          <Badge variant="default" className="text-xs sm:text-sm">
            {totalCount} filtered
          </Badge>
        </div>
      </div>

      {/* Filter Panel */}
      <FilterPanel
        globalFilter={filters.searchTerm}
        onGlobalFilterChange={setSearchTerm}
        searchField={filters.searchField}
        onSearchFieldChange={setSearchField}
        sorting={sorting}
        onSortingChange={setSorting}
        sortField={filters.sortField}
        onSortFieldChange={setSortField}
        sortDirection={filters.sortDirection}
        onSortDirectionChange={setSortDirection}
        columnFilters={columnFilters}
        onColumnFiltersChange={setColumnFilters}
        statusFilter={filters.statusFilter}
        onStatusFilterChange={setStatusFilter}
        serviceFilter={filters.serviceFilter}
        onServiceFilterChange={setServiceFilter}
        teamFilter={filters.teamFilter}
        onTeamFilterChange={setTeamFilter}
        totalCount={totalCount}
        filteredCount={filteredData.length}
        onClearAll={onClearAll}
      />

      {/* Data Table */}
      <Card>
        <CardHeader className="pb-4">
          <CardTitle className="text-lg sm:text-xl">Flag Submissions</CardTitle>
          <CardDescription className="text-sm"></CardDescription>
        </CardHeader>
        <CardContent className="p-0 sm:p-6">
          <DataTable
            key={`${JSON.stringify(filters)}-${pagination.pageIndex}-${pagination.pageSize}`}
            columns={columns}
            data={filteredData}
            totalCount={totalCount}
            pageIndex={pagination.pageIndex}
            pageSize={pagination.pageSize}
            onPaginationChange={setPagination}
          />
        </CardContent>
      </Card>
    </div>
  );
}
