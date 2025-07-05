import { useState, useMemo, useEffect, useRef, useCallback } from 'react';
import useSWR, { mutate } from 'swr';
import { fetcher } from '@/lib/fetchers';
import { BACKEND_URL } from '@/lib/constants';
import { SortingState, ColumnFiltersState } from '@tanstack/react-table';
import { Flag } from '@/lib/types';

interface PaginationState {
  pageIndex: number;
  pageSize: number;
}

interface FilterState {
  searchTerm: string;
  searchField: string;
  statusFilter: string;
  serviceFilter: string;
  teamFilter: string;
  sortField: string;
  sortDirection: 'asc' | 'desc';
}

interface ApiResponse {
  flags: Flag[];
  n_flags: number;
}

interface UsePaginatedFlagsReturn {
  data: Flag[];
  totalCount: number;
  isLoading: boolean;
  error: any; //eslint-disable-line
  pagination: PaginationState;
  filters: FilterState;
  sorting: SortingState;
  columnFilters: ColumnFiltersState;
  setPagination: (pagination: PaginationState) => void;
  setFilters: (filters: Partial<FilterState>) => void;
  setSearchTerm: (term: string) => void;
  setSearchField: (field: string) => void;
  setStatusFilter: (status: string) => void;
  setServiceFilter: (service: string) => void;
  setTeamFilter: (team: string) => void;
  setSortField: (field: string) => void;
  setSortDirection: (direction: 'asc' | 'desc') => void;
  setSorting: (sorting: SortingState) => void;
  setColumnFilters: (filters: ColumnFiltersState) => void;
  onClearAll: () => void;
  forceRefetch: () => void;
}

export function usePaginatedFlags(): UsePaginatedFlagsReturn {
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 20,
  });

  const [filters, setFiltersState] = useState<FilterState>({
    searchTerm: '',
    searchField: 'all',
    statusFilter: 'all',
    serviceFilter: '',
    teamFilter: '',
    sortField: 'submit_time',
    sortDirection: 'desc',
  });

  const [debouncedSearchTerm, setDebouncedSearchTerm] = useState('');
  const debounceTimer = useRef<NodeJS.Timeout | undefined>(undefined);

  // TanStack Table state
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);

  // Proper debounce for search term
  const debounceSearch = useCallback((searchTerm: string) => {
    if (debounceTimer.current) {
      clearTimeout(debounceTimer.current);
    }

    debounceTimer.current = setTimeout(() => {
      setDebouncedSearchTerm(searchTerm);
    }, 300);
  }, []);

  useEffect(() => {
    debounceSearch(filters.searchTerm);
    return () => {
      if (debounceTimer.current) {
        clearTimeout(debounceTimer.current);
      }
    };
  }, [filters.searchTerm, debounceSearch]);

  // Build URL with new clean endpoint parameters
  const buildUrl = () => {
    const baseUrl = `${BACKEND_URL}/api/v1/flags/${pagination.pageSize}`;
    const params = new URLSearchParams();

    params.set(
      'offset',
      (pagination.pageIndex * pagination.pageSize).toString(),
    );

    // Status filter
    if (filters.statusFilter !== 'all') {
      params.set('status', filters.statusFilter);
    }

    // Service filter
    if (filters.serviceFilter) {
      params.set('service', filters.serviceFilter);
    }

    // Team filter
    if (filters.teamFilter) {
      params.set('team', filters.teamFilter);
    }

    // Search parameters
    if (debouncedSearchTerm) {
      params.set('search', debouncedSearchTerm);
      params.set('search_field', filters.searchField);
    }

    // Sort parameters
    if (filters.sortField) {
      params.set('sort_field', filters.sortField);
      params.set('sort_dir', filters.sortDirection.toUpperCase());
    }
    const finalUrl = `${baseUrl}?${params.toString()}`;
    return finalUrl;
  };

  const paginatedUrl = buildUrl();
  const {
    data: rawData,
    error,
    isLoading,
  } = useSWR<ApiResponse>(paginatedUrl, fetcher, {
    revalidateOnFocus: false,
    revalidateOnReconnect: true,
    dedupingInterval: 0,
    keepPreviousData: true,
    refreshInterval: 30_000,
  });

  const filteredData = useMemo(() => {
    if (!rawData?.flags || !Array.isArray(rawData.flags)) {
      return [];
    }
    return rawData.flags;
  }, [rawData?.flags]);

  const setFilters = (newFilters: Partial<FilterState>) => {
    setFiltersState(prev => ({ ...prev, ...newFilters }));
    // Reset to first page when filters change
    setPagination(prev => ({ ...prev, pageIndex: 0 }));
  };

  const setSearchTerm = (searchTerm: string) => {
    setFilters({ searchTerm });
  };

  const setSearchField = (searchField: string) => {
    setFilters({ searchField });
  };

  const setStatusFilter = (statusFilter: string) => {
    setFilters({ statusFilter });
  };

  const setServiceFilter = (serviceFilter: string) => {
    setFilters({ serviceFilter });
  };

  const setTeamFilter = (teamFilter: string) => {
    setFilters({ teamFilter });
  };

  const setSortField = (sortField: string) => {
    setFilters({ sortField });
  };

  const setSortDirection = (sortDirection: 'asc' | 'desc') => {
    setFilters({ sortDirection });
  };

  const onClearAll = () => {
    setFiltersState({
      searchTerm: '',
      searchField: 'all',
      statusFilter: 'all',
      serviceFilter: '',
      teamFilter: '',
      sortField: 'submit_time',
      sortDirection: 'desc',
    });
    setSorting([]);
    setColumnFilters([]);
    setPagination(prev => ({ ...prev, pageIndex: 0 }));
  };

  const forceRefetch = () => {
    mutate(paginatedUrl);
  };

  // Reset to first page when filtering changes
  useEffect(() => {
    setPagination(prev => ({ ...prev, pageIndex: 0 }));
  }, [
    filters.searchTerm,
    filters.statusFilter,
    filters.serviceFilter,
    filters.teamFilter,
    filters.sortField,
    filters.sortDirection,
    filters.searchField,
  ]);

  return {
    data: filteredData,
    totalCount: rawData?.n_flags || 0,
    isLoading: isLoading,
    error,
    pagination,
    filters,
    sorting,
    columnFilters,
    setPagination,
    setFilters,
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
    forceRefetch,
  };
}
