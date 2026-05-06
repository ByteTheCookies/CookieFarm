import { useDeferredValue, useMemo, useRef, useState } from "react";
import { ClipboardText } from "@cloudflare/kumo/components/clipboard-text";
import { Empty } from "@cloudflare/kumo/components/empty";
import { Table } from "@cloudflare/kumo/components/table";
import { CaretDownIcon, CaretUpIcon, CaretUpDownIcon } from "@phosphor-icons/react";
import {
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import type {
  ColumnDef,
  ColumnFiltersState,
  Row,
  SortingState,
} from "@tanstack/react-table";
import { useVirtualizer } from "@tanstack/react-virtual";
import type { Flag, FlagStatus } from "@/api/flags";
import { FlagStatusBadge } from "@/components/FlagStatusBadge";

const flagStatusLabels = {
  0: "Queued",
  1: "Accepted",
  2: "Denied",
  3: "Error",
  4: "Invalid",
} satisfies Record<FlagStatus, string>;

function formatTimestamp(timestamp: number): string {
  if (timestamp === 0) {
    return "-";
  }

  return new Date(timestamp * 1000).toLocaleString([], {
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function formatDuration(flag: Flag): string {
  if (flag.response_time === 0 || flag.response_time < flag.submit_time) {
    return "Pending";
  }

  return `${flag.response_time - flag.submit_time}s`;
}

function getFlagRowId(flag: Flag): string {
  return `${flag.flag_code}-${flag.submit_time}`;
}

function SortIcon(props: { direction: false | "asc" | "desc" }) {
  if (props.direction === "asc") {
    return <CaretUpIcon aria-hidden="true" size={13} weight="bold" />;
  }

  if (props.direction === "desc") {
    return <CaretDownIcon aria-hidden="true" size={13} weight="bold" />;
  }

  return <CaretUpDownIcon aria-hidden="true" size={13} />;
}

export function FlagTable(props: Readonly<{
  rows: Flag[];
  emptyTitle?: string;
  emptyDescription?: string;
}>) {
  const parentRef = useRef<HTMLDivElement | null>(null);
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnSizing, setColumnSizing] = useState({});
  const [globalFilterInput] = useState("");
  const globalFilter = useDeferredValue(globalFilterInput);
  const columns = useMemo<ColumnDef<Flag>[]>(
    () => [
      {
        id: "flag_code",
        accessorKey: "flag_code",
        header: "Flag",
        size: 320,
        minSize: 220,
        cell: ({ getValue }) => (
          <ClipboardText size="sm" text={getValue<string>()} />
        ),
      },
      {
        id: "msg",
        accessorKey: "msg",
        header: "Message",
        size: 360,
        minSize: 180,
        cell: ({ getValue }) => (
          <span className="line-clamp-2 wrap-break-words">{getValue<string>() || "-"}</span>
        ),
      },
      {
        id: "service",
        accessorFn: (flag) => `${flag.service_name}:${flag.port_service}`,
        header: "Service",
        size: 160,
        minSize: 120,
        cell: ({ getValue }) => (
          <span className="whitespace-nowrap">{getValue<string>()}</span>
        ),
      },
      {
        id: "status",
        accessorFn: (flag) => flagStatusLabels[flag.status],
        header: "Status",
        size: 120,
        minSize: 100,
        cell: ({ row }) => <FlagStatusBadge status={row.original.status} />,
        filterFn: "equalsString",
      },
      {
        id: "submit_time",
        accessorKey: "submit_time",
        header: "Submit",
        size: 184,
        minSize: 150,
        cell: ({ getValue }) => formatTimestamp(getValue<number>()),
      },
      {
        id: "response_time",
        accessorKey: "response_time",
        header: "Response",
        size: 184,
        minSize: 150,
        cell: ({ getValue }) => formatTimestamp(getValue<number>()),
      },
      {
        id: "duration",
        accessorFn: formatDuration,
        header: "Duration",
        size: 96,
        minSize: 88,
        cell: ({ getValue }) => (
          <span className="block text-center">{getValue<string>()}</span>
        ),
        sortingFn: (rowA, rowB) => {
          const getDuration = (row: Row<Flag>) =>
            row.original.response_time === 0 ||
              row.original.response_time < row.original.submit_time
              ? Number.POSITIVE_INFINITY
              : row.original.response_time - row.original.submit_time;

          return getDuration(rowA) - getDuration(rowB);
        },
      },
      {
        id: "team_id",
        accessorKey: "team_id",
        header: "Team",
        size: 80,
        minSize: 72,
        cell: ({ getValue }) => (
          <span className="block text-center">{getValue<number>()}</span>
        ),
      },
    ],
    [],
  );
  const table = useReactTable({
    data: props.rows,
    columns,
    state: {
      sorting,
      columnFilters,
      columnSizing,
      globalFilter,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnSizingChange: setColumnSizing,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getRowId: getFlagRowId,
    columnResizeMode: "onChange",
    globalFilterFn: "includesString",
  });
  const tableRows = table.getRowModel().rows;
  const shouldVirtualize = tableRows.length > 500;
  const virtualizer = useVirtualizer({
    count: tableRows.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 54,
    overscan: 12,
  });

  if (props.rows.length === 0) {
    return (
      <section className="rounded-2xl border border-kumo-line bg-kumo-base">
        <Empty
          size="sm"
          title={props.emptyTitle ?? "No flags match these filters"}
          description={
            props.emptyDescription ??
            "Adjust the current filters to widen the view or wait for new rows."
          }
        />
      </section>
    );
  }

  const virtualRows = shouldVirtualize ? virtualizer.getVirtualItems() : [];
  const paddingTop = shouldVirtualize && virtualRows[0] ? virtualRows[0].start : 0;
  const paddingBottom =
    shouldVirtualize && virtualRows.length > 0
      ? virtualizer.getTotalSize() - virtualRows[virtualRows.length - 1]!.end
      : 0;
  const visibleRows: Row<Flag>[] = shouldVirtualize
    ? virtualRows.flatMap((item) => {
      const row = tableRows[item.index];
      return row ? [row] : [];
    })
    : tableRows;
  const columnCount = table.getVisibleLeafColumns().length;
  return (
    <section className="overflow-hidden rounded-2xl border border-kumo-line bg-kumo-base">
      <div ref={parentRef} className="max-h-[70vh] overflow-auto">
        <Table layout="fixed">
          <colgroup>
            {table.getVisibleLeafColumns().map((column) => (
              <col key={column.id} style={{ width: column.getSize() }} />
            ))}
          </colgroup>
          <Table.Header sticky>
            {table.getHeaderGroups().map((headerGroup) => (
              <Table.Row key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  const sortedDirection = header.column.getIsSorted();

                  return (
                    <Table.Head key={header.id} className="relative pr-2">
                      {header.isPlaceholder ? null : (
                        <button
                          type="button"
                          className="flex w-full items-center justify-between gap-2 text-left disabled:cursor-default"
                          disabled={!header.column.getCanSort()}
                          onClick={header.column.getToggleSortingHandler()}
                          aria-sort={
                            sortedDirection === "asc"
                              ? "ascending"
                              : sortedDirection === "desc"
                                ? "descending"
                                : "none"
                          }
                        >
                          <span className="truncate">
                            {flexRender(
                              header.column.columnDef.header,
                              header.getContext(),
                            )}
                          </span>
                          {header.column.getCanSort() ? (
                            <SortIcon direction={sortedDirection} />
                          ) : null}
                        </button>
                      )}
                      {header.column.getCanResize() ? (
                        <Table.ResizeHandle
                          aria-label={`Resize ${header.column.id} column`}
                          onMouseDown={header.getResizeHandler()}
                          onTouchStart={header.getResizeHandler()}
                        />
                      ) : null}
                    </Table.Head>
                  );
                })}
              </Table.Row>
            ))}
          </Table.Header>
          <Table.Body>
            {paddingTop > 0 ? (
              <Table.Row aria-hidden="true">
                <Table.Cell colSpan={columnCount} style={{ height: paddingTop }} />
              </Table.Row>
            ) : null}

            {visibleRows.map((row) => (
              <Table.Row key={row.id}>
                {row.getVisibleCells().map((cell) => (
                  <Table.Cell key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </Table.Cell>
                ))}
              </Table.Row>
            ))}

            {tableRows.length === 0 ? (
              <Table.Row>
                <Table.Cell colSpan={columnCount} className="py-10 text-center">
                  <Empty
                    size="sm"
                    title="No visible flags match this table filter"
                    description="Clear the table filter or reset sorting and filters."
                  />
                </Table.Cell>
              </Table.Row>
            ) : null}

            {paddingBottom > 0 ? (
              <Table.Row aria-hidden="true">
                <Table.Cell colSpan={columnCount} style={{ height: paddingBottom }} />
              </Table.Row>
            ) : null}
          </Table.Body>
        </Table>
      </div>
    </section>
  );
}
