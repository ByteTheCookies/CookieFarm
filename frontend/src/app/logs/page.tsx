'use client';

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Flag, getColumns } from './column';
import { DataTable } from './data-table';
import { FilterPanel } from './filter-panel';
import { usePaginatedFlags } from '@/hooks/usePaginatedFlags';
import axios from 'axios';
import { BACKEND_URL } from '@/lib/constants';
import { RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';

export default function FlagLogs() {
  const {
    data: filteredData,
    totalCount,
    error,
    pagination,
    filters,
    sorting,
    columnFilters,
    isLoading,
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
    forceRefetch,
    onClearAll,
  } = usePaginatedFlags();

  function deleteFlag(flag: Flag) {
    axios
      .delete(`${BACKEND_URL}/api/v1/delete-flag?flag=${flag.flag_code}`, {
        withCredentials: true,
      })
      .then(() => {
        toast.success('Flag deleted successfully');
        forceRefetch();
      })
      .catch(error => {
        toast.error('Error deleting flag');
        console.error(error);
        alert('Error deleting flag');
      });
  }

  function copyFlagCode(flag: Flag) {
    navigator.clipboard
      .writeText(flag.flag_code)
      .then(() => {
        toast.success('Flag copied to clipboard');
      })
      .catch(error => {
        toast.error('Error copying flag');
        console.error(error);
        alert('Error copying flag');
      });
  }

  function submitFlag(flag: Flag) {
    axios
      .post(
        `${BACKEND_URL}/api/v1/submit-flag`,
        { flag: flag },
        {
          headers: {
            'Content-Type': 'application/json',
          },
          withCredentials: true,
        },
      )
      .then(() => {
        toast.success('Flag submitted successfully');
        forceRefetch();
      })
      .catch(error => {
        toast.error('Error submitting flag');
        console.error(error);
        alert('Error submitting flag');
      });
  }

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
    <div className="container mx-auto space-y-6 px-8 py-12">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight sm:text-3xl">
            Flag Table
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
          {/* Reload button */}
          <Button
            variant="outline"
            className="flex items-center gap-1 text-xs sm:text-sm"
            onClick={forceRefetch}
            disabled={isLoading}
          >
            {isLoading ? (
              <RefreshCw className="h-4 w-4 animate-spin" />
            ) : (
              <RefreshCw className="h-4 w-4" />
            )}
            <span>Reload</span>
          </Button>
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
        <CardHeader className="ml-2 flex items-center gap-3 pb-2">
          <CardTitle className="text-lg sm:text-2xl">
            Flag Submissions
          </CardTitle>
          <CardDescription className="text-sm">
            Eat all the cookies
          </CardDescription>
        </CardHeader>
        <CardContent className="p-0 sm:px-6">
          <DataTable
            key={`${JSON.stringify(filters)}-${pagination.pageIndex}-${pagination.pageSize}`}
            columns={getColumns(deleteFlag, submitFlag, copyFlagCode)}
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
