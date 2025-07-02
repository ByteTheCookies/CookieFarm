'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
// Removed Collapsible import for compatibility
import { Badge } from '@/components/ui/badge';
import {
  Search,
  Filter,
  X,
  ChevronDown,
  ChevronUp,
  SortAsc,
  SortDesc,
  ArrowUpDown,
} from 'lucide-react';
import { SortingState, ColumnFiltersState } from '@tanstack/react-table';

interface FilterPanelProps {
  // Search
  globalFilter: string;
  onGlobalFilterChange: (value: string) => void;
  searchField: string;
  onSearchFieldChange: (field: string) => void;

  // Sorting
  sorting: SortingState;
  onSortingChange: (sorting: SortingState) => void;
  sortField: string;
  onSortFieldChange: (field: string) => void;
  sortDirection: 'asc' | 'desc';
  onSortDirectionChange: (direction: 'asc' | 'desc') => void;

  // Filters
  columnFilters: ColumnFiltersState;
  onColumnFiltersChange: (filters: ColumnFiltersState) => void;
  statusFilter: string;
  onStatusFilterChange: (status: string) => void;
  serviceFilter: string;
  onServiceFilterChange: (service: string) => void;
  teamFilter: string;
  onTeamFilterChange: (team: string) => void;

  // Stats
  totalCount: number;
  filteredCount: number;

  // Actions
  onClearAll: () => void;
}

const searchableFields = [
  { value: 'all', label: 'All Fields' },
  { value: 'flag_code', label: 'Flag Code' },
  { value: 'service_name', label: 'Service Name' },
  { value: 'port_service', label: 'Service Port' },
  { value: 'exploit_name', label: 'Exploit Name' },
  { value: 'msg', label: 'Message' },
  { value: 'username', label: 'Username' },
];

const sortableFields = [
  { value: 'submit_time', label: 'Submit Time' },
  { value: 'response_time', label: 'Response Time' },
  { value: 'status', label: 'Status' },
  { value: 'flag_code', label: 'Flag Code' },
  { value: 'service_name', label: 'Service Name' },
  { value: 'team_id', label: 'Team ID' },
  { value: 'port_service', label: 'Port' },
];

const statusOptions = [
  { value: 'all', label: 'All Statuses' },
  { value: 'ACCEPTED', label: 'Accepted' },
  { value: 'DENIED', label: 'Denied' },
  { value: 'RESUBMIT', label: 'Resubmit' },
  { value: 'ERROR', label: 'Error' },
];

export function FilterPanel({
  globalFilter,
  onGlobalFilterChange,
  searchField,
  onSearchFieldChange,
  sorting,
  sortField,
  onSortFieldChange,
  sortDirection,
  onSortDirectionChange,
  columnFilters,
  statusFilter,
  onStatusFilterChange,
  serviceFilter,
  onServiceFilterChange,
  teamFilter,
  onTeamFilterChange,
  totalCount,
  filteredCount,
  onClearAll,
}: FilterPanelProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const hasActiveFilters =
    globalFilter ||
    statusFilter !== 'all' ||
    serviceFilter ||
    teamFilter ||
    columnFilters.length > 0;

  const activeFiltersCount = [
    globalFilter,
    statusFilter !== 'all' ? statusFilter : null,
    serviceFilter,
    teamFilter,
    ...columnFilters.map(f => f.value),
  ].filter(Boolean).length;

  return (
    <Card className="w-full">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2 text-lg">
              <Filter className="h-5 w-5" />
              Search & Filters
              {activeFiltersCount > 0 && (
                <Badge variant="secondary" className="ml-2">
                  {activeFiltersCount}
                </Badge>
              )}
            </CardTitle>
            <CardDescription className="text-sm">
              {filteredCount} of {totalCount} flags shown
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            {hasActiveFilters && (
              <Button
                variant="outline"
                size="sm"
                onClick={onClearAll}
                className="h-8"
              >
                <X className="mr-1 h-4 w-4" />
                Clear All
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              className="h-8 w-8 p-0"
              onClick={() => setIsExpanded(!isExpanded)}
            >
              {isExpanded ? (
                <ChevronUp className="h-4 w-4" />
              ) : (
                <ChevronDown className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Quick Search Row - Always Visible */}
        <div className="grid grid-cols-1 gap-3 md:grid-cols-12">
          {/* Search Field Selector */}
          <div className="flex gap-2 md:col-span-3">
            <Select value={searchField} onValueChange={onSearchFieldChange}>
              <SelectTrigger className="h-10">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {searchableFields.map(field => (
                  <SelectItem key={field.value} value={field.value}>
                    {field.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* Quick Status Filter */}
            <Select value={statusFilter} onValueChange={onStatusFilterChange}>
              <SelectTrigger className="h-10">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {statusOptions.map(option => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Search Input */}
          <div className="relative md:col-span-6">
            <Search className="text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 transform" />
            <Input
              placeholder={`Search ${searchableFields.find(f => f.value === searchField)?.label.toLowerCase() || 'all fields'}...`}
              value={globalFilter}
              onChange={e => onGlobalFilterChange(e.target.value)}
              className="h-10 pl-10"
            />
            {globalFilter && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onGlobalFilterChange('')}
                className="absolute top-1/2 right-2 h-6 w-6 -translate-y-1/2 transform p-0"
              >
                <X className="h-3 w-3" />
              </Button>
            )}
          </div>
        </div>

        {/* Expanded Filters */}
        {isExpanded && (
          <div className="space-y-4 border-t pt-4">
            {/* Sort Controls */}
            <div className="space-y-2">
              <h4 className="flex items-center gap-2 text-sm font-medium">
                <ArrowUpDown className="h-4 w-4" />
                Sort By
              </h4>
              <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                <Select value={sortField} onValueChange={onSortFieldChange}>
                  <SelectTrigger>
                    <SelectValue placeholder="Select field to sort by" />
                  </SelectTrigger>
                  <SelectContent>
                    {sortableFields.map(field => (
                      <SelectItem key={field.value} value={field.value}>
                        {field.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                <Select
                  value={sortDirection}
                  onValueChange={onSortDirectionChange}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="asc">
                      <div className="flex items-center gap-2">
                        <SortAsc className="h-4 w-4" />
                        Ascending
                      </div>
                    </SelectItem>
                    <SelectItem value="desc">
                      <div className="flex items-center gap-2">
                        <SortDesc className="h-4 w-4" />
                        Descending
                      </div>
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>

            {/* Active Filters Display */}
            {hasActiveFilters && (
              <div className="space-y-2">
                <h4 className="text-sm font-medium">Active Filters</h4>
                <div className="flex flex-wrap gap-2">
                  {globalFilter && (
                    <Badge variant="secondary" className="text-xs">
                      Search: &quot;{globalFilter}&quot;
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onGlobalFilterChange('')}
                        className="ml-1 h-auto p-0 text-xs hover:bg-transparent"
                      >
                        <X className="h-3 w-3" />
                      </Button>
                    </Badge>
                  )}

                  {statusFilter !== 'all' && (
                    <Badge variant="secondary" className="text-xs">
                      Status:{' '}
                      {statusOptions.find(s => s.value === statusFilter)?.label}
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onStatusFilterChange('all')}
                        className="ml-1 h-auto p-0 text-xs hover:bg-transparent"
                      >
                        <X className="h-3 w-3" />
                      </Button>
                    </Badge>
                  )}

                  {serviceFilter && (
                    <Badge variant="secondary" className="text-xs">
                      Service: &quot;{serviceFilter}&quot;
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onServiceFilterChange('')}
                        className="ml-1 h-auto p-0 text-xs hover:bg-transparent"
                      >
                        <X className="h-3 w-3" />
                      </Button>
                    </Badge>
                  )}

                  {teamFilter && (
                    <Badge variant="secondary" className="text-xs">
                      Team: {teamFilter}
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onTeamFilterChange('')}
                        className="ml-1 h-auto p-0 text-xs hover:bg-transparent"
                      >
                        <X className="h-3 w-3" />
                      </Button>
                    </Badge>
                  )}

                  {sorting.length > 0 && (
                    <Badge variant="outline" className="text-xs">
                      Sort:{' '}
                      {sortableFields.find(f => f.value === sortField)?.label} (
                      {sortDirection})
                    </Badge>
                  )}
                </div>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
