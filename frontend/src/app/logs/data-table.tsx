'use client';

import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table';

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  totalCount: number;
  pageIndex: number;
  pageSize: number;
  onPaginationChange: (pagination: {
    pageIndex: number;
    pageSize: number;
  }) => void;
}

export function DataTable<TData, TValue>({
  columns,
  data,
  totalCount,
  pageIndex,
  pageSize,
  onPaginationChange,
}: DataTableProps<TData, TValue>) {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    pageCount: Math.ceil(totalCount / pageSize),
    state: {
      pagination: {
        pageIndex,
        pageSize,
      },
    },
    onPaginationChange: updater => {
      if (typeof updater === 'function') {
        const newState = updater({ pageIndex, pageSize });
        onPaginationChange(newState);
      } else {
        onPaginationChange(updater);
      }
    },
  });

  return (
    <div className="space-y-4">
      <div className="overflow-hidden rounded-md border">
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              {table.getHeaderGroups().map(headerGroup => (
                <TableRow key={headerGroup.id} className="h-16">
                  {headerGroup.headers.map(header => {
                    return (
                      <TableHead key={header.id} className="font-bold">
                        {header.isPlaceholder
                          ? null
                          : flexRender(
                              header.column.columnDef.header,
                              header.getContext(),
                            )}
                      </TableHead>
                    );
                  })}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {table.getRowModel().rows?.length ? (
                table.getRowModel().rows.map(row => (
                  <TableRow
                    className="h-16"
                    key={row.id}
                    data-state={row.getIsSelected() && 'selected'}
                  >
                    {row.getVisibleCells().map(cell => (
                      <TableCell key={cell.id}>
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext(),
                        )}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell
                    colSpan={columns.length}
                    className="h-32 text-center"
                  >
                    <div className="flex flex-col items-center space-y-2">
                      <div className="text-muted-foreground">
                        {totalCount === 0 ? (
                          <>
                            <p className="text-lg font-medium">
                              No flags found
                            </p>
                            <p className="text-sm">
                              There are no flags in the system yet.
                            </p>
                          </>
                        ) : (
                          <>
                            <p className="text-lg font-medium">
                              No matching results
                            </p>
                            <p className="text-sm">
                              Try adjusting your search criteria or filters.
                            </p>
                          </>
                        )}
                      </div>
                    </div>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
        <div className="mx-4 flex flex-col gap-4 py-4 sm:flex-row sm:items-center sm:justify-between">
          {/* Page Size Selector */}
          <div className="flex items-center space-x-2">
            <p className="text-xs font-medium whitespace-nowrap sm:text-sm">
              Rows per page
            </p>
            <Select
              value={`${pageSize}`}
              onValueChange={value => {
                onPaginationChange({
                  pageIndex: 0,
                  pageSize: Number(value),
                });
              }}
            >
              <SelectTrigger className="h-8 w-[60px] sm:w-[70px]">
                <SelectValue placeholder={pageSize} />
              </SelectTrigger>
              <SelectContent side="top">
                {[20, 50, 100, 200].map(size => (
                  <SelectItem key={size} value={`${size}`}>
                    {size}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Page Info */}
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-6">
            <div className="text-center sm:text-left">
              <p className="text-xs font-medium sm:text-sm">
                Page {pageIndex + 1} of {Math.ceil(totalCount / pageSize)}
              </p>
              <p className="text-muted-foreground text-xs">
                Showing {pageIndex * pageSize + 1} to{' '}
                {Math.min((pageIndex + 1) * pageSize, totalCount)} of{' '}
                {totalCount}
              </p>
            </div>

            {/* Navigation Controls */}
            <div className="flex items-center justify-center space-x-1 sm:space-x-2">
              <Button
                variant="outline"
                size="sm"
                className="hidden h-8 w-8 p-0 sm:flex"
                onClick={() => onPaginationChange({ pageIndex: 0, pageSize })}
                disabled={pageIndex === 0}
              >
                <span className="sr-only">Go to first page</span>
                {'<<'}
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="h-8 w-8 p-0"
                onClick={() => table.previousPage()}
                disabled={!table.getCanPreviousPage()}
              >
                <span className="sr-only">Go to previous page</span>
                {'<'}
              </Button>

              {/* Page Numbers for larger screens */}
              <div className="hidden items-center space-x-1 md:flex">
                {Array.from(
                  { length: Math.min(5, Math.ceil(totalCount / pageSize)) },
                  (_, i) => {
                    const pageNum =
                      Math.max(
                        0,
                        Math.min(
                          Math.ceil(totalCount / pageSize) - 5,
                          pageIndex - 2,
                        ),
                      ) + i;

                    if (pageNum >= Math.ceil(totalCount / pageSize))
                      return null;

                    return (
                      <Button
                        key={pageNum}
                        variant={pageNum === pageIndex ? 'default' : 'outline'}
                        size="sm"
                        className="h-8 w-8 p-0"
                        onClick={() =>
                          onPaginationChange({ pageIndex: pageNum, pageSize })
                        }
                      >
                        {pageNum + 1}
                      </Button>
                    );
                  },
                )}
              </div>

              <Button
                variant="outline"
                size="sm"
                className="h-8 w-8 p-0"
                onClick={() => table.nextPage()}
                disabled={!table.getCanNextPage()}
              >
                <span className="sr-only">Go to next page</span>
                {'>'}
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="hidden h-8 w-8 p-0 sm:flex"
                onClick={() =>
                  onPaginationChange({
                    pageIndex: Math.ceil(totalCount / pageSize) - 1,
                    pageSize,
                  })
                }
                disabled={pageIndex >= Math.ceil(totalCount / pageSize) - 1}
              >
                <span className="sr-only">Go to last page</span>
                {'>>'}
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
