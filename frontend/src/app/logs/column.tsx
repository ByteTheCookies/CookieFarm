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

import {
  Tooltip,
  TooltipProvider,
  TooltipTrigger,
  TooltipContent,
} from '@/components/ui/tooltip';

import {
  Check,
  Clock,
  AlertCircle,
  Ban,
  AlertTriangle,
  MoreHorizontal,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';

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
        <div className="text-center">
          <TooltipProvider>
            <Tooltip bg->
              <TooltipTrigger asChild className="items-center">
                <Badge className="border-green-500/60 bg-green-500/10 p-1 text-green-500">
                  <Check />
                </Badge>
              </TooltipTrigger>
              <TooltipContent className="bg-green-500/20 text-green-500">
                <p>Accepted</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      );
    case 'RESUBMIT':
      return (
        <div className="text-center">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild className="items-center">
                <Badge className="border-yellow-500/60 bg-yellow-500/10 p-1 text-yellow-500">
                  <AlertTriangle className="h-4 w-4" />
                </Badge>
              </TooltipTrigger>
              <TooltipContent className="bg-yellow-500/20 text-yellow-500">
                <p>Resubmit</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      );

    case 'DENIED':
      return (
        <div className="text-center">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild className="items-center">
                <Badge className="border-red-500/60 bg-red-500/10 p-1 text-red-500">
                  <Ban className="h-4 w-4" />
                </Badge>
              </TooltipTrigger>
              <TooltipContent className="bg-red-500/20 text-red-500">
                <p>Denied</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      );

    case 'ERROR':
      return (
        <div className="text-center">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild className="items-center">
                <Badge className="border-purple-500/60 bg-purple-500/10 p-1 text-purple-500">
                  <AlertCircle className="h-4 w-4" />
                </Badge>
              </TooltipTrigger>
              <TooltipContent className="bg-purple-500/20 text-purple-500">
                <p>Error</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      );

    default:
      // NEW: Pending state (default)
      return (
        <div className="text-center">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild className="items-center">
                <Badge className="border-gray-500/60 bg-gray-500/10 p-1 text-gray-500">
                  <Clock className="h-4 w-4" />
                </Badge>
              </TooltipTrigger>
              <TooltipContent className="bg-gray-500/10 text-gray-500">
                <p>Pending</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      );
  }
};

export const columns: ColumnDef<Flag>[] = [
  {
    accessorKey: 'status',
    header: ' ',
    cell: ({ row }) => {
      return getStatusBadge(row.getValue('status'));
    },
    size: 140,
  },
  {
    accessorKey: 'flag_code',
    header: 'Flag Code',
    cell: ({ row }) => {
      const flagCode = row.getValue('flag_code') as string;
      return (
        <div
          className="max-w-[140px] font-mono text-xs text-white sm:max-w-[180px] sm:text-sm"
          title={flagCode}
        >
          {flagCode}
        </div>
      );
    },
    size: 180,
  },
  {
    accessorKey: 'msg',
    header: 'Message',
    cell: ({ row }) => {
      const message = row.getValue('msg') as string;
      return (
        <div
          className="w-[180px] max-w-[300px] truncate text-xs font-light text-gray-200 sm:max-w-[300px] sm:text-sm"
          title={message}
        >
          {message}
        </div>
      );
    },
    size: 300,
  },
  {
    header: 'Service',
    accessorFn: row => `${row.service_name}:${row.port_service}`,
    cell: ({ row }) => {
      const serviceName = row.original.service_name as string;
      const servicePort = row.original.port_service as number;
      return (
        <div
          className="max-w-[100px] truncate sm:max-w-[150px]"
          title={`${serviceName}:${servicePort}`}
        >
          <p className="text-gray-200">{serviceName}</p>
          <p className="text-muted-foreground text-xs">{servicePort}</p>
        </div>
      );
    },
    size: 150,
  },
  {
    header: 'Victim',
    accessorFn: row => `${row.team_id} with ${row.exploit_name}`,
    cell: ({ row }) => {
      const exploitName = row.original.exploit_name as string;
      const teamId = row.original.team_id as number;

      return (
        <div
          className="max-w-[100px] truncate sm:max-w-[150px]"
          title={exploitName}
        >
          <p className="text-sm text-gray-200">
            Team <span className="font-mono">{teamId}</span>
          </p>
          <p className="text-muted-foreground font-mono text-xs">
            with <span className="text-rose-400">{exploitName}</span>
          </p>
        </div>
      );
    },
    size: 150,
  },
  {
    accessorKey: 'submit_time',
    header: 'Submitted',
    cell: ({ row }) => {
      const submitTime = row.getValue('submit_time') as number;
      const date = new Date(submitTime * 1000);
      return (
        <div className="text-sm whitespace-nowrap text-gray-200 sm:text-sm">
          <div>{date.toLocaleTimeString()}</div>
          <div className="text-muted-foreground text-xs">
            {date.toLocaleDateString()}
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
        <Badge variant="outline" className="font-mono text-xs text-gray-200">
          {duration} s
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
