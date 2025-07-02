'use client';

import { ColumnDef } from '@tanstack/react-table';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Badge } from '@/components/ui/badge';
import { MoreHorizontal, ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react';

// This type is used to define the shape of our data.
// You can use a Zod schema here if you want.
export type Flag = {
  flag_code: string;
  service_name: string;
  status: 'ACCEPTED' | 'DENIED' | 'RESUBMIT' | 'ERROR';
  exploit_name: string;
  msg: string;
  submit_time: number;
  response_time: number;
  port_service: number;
  team_id: number;
};

const getStatusBadge = (status: string) => {
  switch (status) {
    case 'ACCEPTED':
      return (
        <Badge className="border-green-500/20 bg-green-500/10 text-green-500">
          Accepted
        </Badge>
      );
    case 'RESUBMIT':
      return (
        <Badge className="border-yellow-500/20 bg-yellow-500/10 text-yellow-500">
          Resubmit
        </Badge>
      );
    case 'DENIED':
      return (
        <Badge className="border-red-500/20 bg-red-500/10 text-red-500">
          Denied
        </Badge>
      );
    case 'ERROR':
      return (
        <Badge className="border-gray-500/20 bg-gray-500/10 text-gray-500">
          Error
        </Badge>
      );
    default:
      return <Badge variant="outline">{status}</Badge>;
  }
};

export const columns: ColumnDef<Flag>[] = [
  {
    accessorKey: 'status',
    header: 'Status',
    cell: ({ row }) => {
      return getStatusBadge(row.getValue('status'));
    },
    size: 120,
  },
  {
    accessorKey: 'flag_code',
    header: 'Flag Code',
    cell: ({ row }) => {
      const flagCode = row.getValue('flag_code') as string;
      return (
        <div
          className="max-w-[120px] truncate font-mono text-xs sm:max-w-[200px] sm:text-sm"
          title={flagCode}
        >
          {flagCode}
        </div>
      );
    },
    size: 200,
  },
  {
    accessorKey: 'service_name',
    header: 'Service',
    cell: ({ row }) => {
      const serviceName = row.getValue('service_name') as string;
      return (
        <div
          className="max-w-[100px] truncate sm:max-w-[150px]"
          title={serviceName}
        >
          {serviceName}
        </div>
      );
    },
    size: 150,
  },
  {
    accessorKey: 'port_service',
    header: 'Port',
    cell: ({ row }) => {
      return (
        <Badge variant="secondary" className="font-mono text-xs">
          {row.getValue('port_service')}
        </Badge>
      );
    },
    size: 80,
  },
  {
    accessorKey: 'exploit_name',
    header: 'Exploit',
    cell: ({ row }) => {
      const exploitName = row.getValue('exploit_name') as string;
      return (
        <div
          className="max-w-[100px] truncate sm:max-w-[150px]"
          title={exploitName}
        >
          {exploitName}
        </div>
      );
    },
    size: 150,
  },
  {
    accessorKey: 'msg',
    header: 'Message',
    cell: ({ row }) => {
      const message = row.getValue('msg') as string;
      return (
        <div
          className="max-w-[150px] truncate text-xs sm:max-w-[300px] sm:text-sm"
          title={message}
        >
          {message}
        </div>
      );
    },
    size: 300,
  },
  {
    accessorKey: 'team_id',
    header: 'Team',
    cell: ({ row }) => {
      return (
        <Badge variant="outline" className="font-mono">
          {row.getValue('team_id')}
        </Badge>
      );
    },
    size: 80,
  },
  {
    accessorKey: 'submit_time',
    header: 'Submitted',
    cell: ({ row }) => {
      const submitTime = row.getValue('submit_time') as number;
      const date = new Date(submitTime * 1000);
      return (
        <div className="text-xs whitespace-nowrap sm:text-sm">
          <div>{date.toLocaleDateString()}</div>
          <div className="text-muted-foreground">
            {date.toLocaleTimeString()}
          </div>
        </div>
      );
    },
    size: 120,
  },
  {
    accessorKey: 'response_time',
    header: 'Duration',
    cell: ({ row }) => {
      const responseTime = row.getValue('response_time') as number;
      const submitTime = row.getValue('submit_time') as number;
      const duration = responseTime - submitTime;
      if (isNaN(duration) || duration < 0) {
        return <span className="text-muted-foreground text-xs">N/A</span>;
      }
      return (
        <Badge variant="outline" className="font-mono text-xs">
          {duration}s
        </Badge>
      );
    },
    size: 80,
  },
  {
    id: 'actions',
    cell: ({ row }) => {
      const flag = row.original;

      return (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="h-8 w-8 p-0">
              <span className="sr-only"> Open menu </span>
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>Actions </DropdownMenuLabel>
            <DropdownMenuItem
              onClick={() => navigator.clipboard.writeText(flag.flag_code)}
            >
              Copy flag code
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem>View details </DropdownMenuItem>
            <DropdownMenuItem> View exploit </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      );
    },
  },
];
