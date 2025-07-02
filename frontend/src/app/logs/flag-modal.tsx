'use client';

import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import {
  Copy,
  X,
  Server,
  Clock,
  Hourglass,
  User,
  Code2,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Trash2,
  Send,
  Eye,
  MoreHorizontal,
  Flag,
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { cn } from '@/lib/utils';

type ModalView = 'details' | 'exploit';

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

interface FlagModalProps {
  isOpen: boolean;
  onClose: () => void;
  flag: Flag;
  onDelete?: () => void;
  onSubmit?: () => void;
  onCopyCode?: () => void;
}

export default function FlagModal({
  isOpen,
  onClose,
  flag,
  onDelete,
  onSubmit,
}: FlagModalProps) {
  const [view, setView] = useState<ModalView>('details');
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'ACCEPTED':
        return <CheckCircle className="h-3 w-3 text-green-400" />;
      case 'DENIED':
        return <XCircle className="h-3 w-3 text-red-400" />;
      case 'PENDING':
        return <Clock className="h-3 w-3 text-yellow-400" />;
      default:
        return <AlertTriangle className="h-3 w-3 text-gray-400" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'ACCEPTED':
        return 'text-green-400';
      case 'DENIED':
        return 'text-red-400';
      case 'PENDING':
        return 'text-yellow-400';
      default:
        return 'text-gray-400';
    }
  };

  const handleCopyCode = () => {
    navigator.clipboard.writeText(flag.flag_code);
  };

  const handleDelete = () => {
    onDelete?.();
    onClose();
  };

  const handleSubmit = () => {
    onSubmit?.();
    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl gap-0 border-[#333333] bg-[#1a1a1a] p-0">
        {/* Header with actions */}

        <DialogTitle className="sr-only">Flag Details</DialogTitle>

        <DialogHeader className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Flag className="w-5 text-[#8b8b8b]" />
              <h2 className="text-lg font-medium text-white">Flag Details</h2>
              <div className="flex items-center gap-1">
                {getStatusIcon(flag.status)}
                <span
                  className={cn(
                    'text-sm font-medium',
                    getStatusColor(flag.status),
                  )}
                >
                  {flag.status}
                </span>
              </div>
            </div>

            <div className="flex items-center gap-2">
              {/* Action Menu */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-[#8b8b8b] hover:bg-[#333333] hover:text-white"
                  >
                    <MoreHorizontal className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="end"
                  className="w-48 border-[#333333] bg-[#2a2a2a] text-white"
                >
                  <DropdownMenuItem
                    onClick={handleSubmit}
                    className="flex items-center gap-2 hover:bg-[#333333] focus:bg-[#333333]"
                  >
                    <Send className="h-4 w-4" />
                    Submit flag
                  </DropdownMenuItem>
                  <DropdownMenuSeparator className="bg-[#333333]" />
                  <DropdownMenuItem
                    onClick={() => setShowDeleteConfirm(true)}
                    className="flex items-center gap-2 text-red-400 hover:bg-red-500/10 focus:bg-red-500/10"
                  >
                    <Trash2 className="h-4 w-4 text-red-400" />
                    Delete flag
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>

              <Button
                variant="ghost"
                size="icon"
                onClick={onClose}
                className="h-8 w-8 text-[#8b8b8b] hover:bg-[#333333] hover:text-white"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          </div>

          {/* Tab Navigation - matching table style */}
          <div className="mt-4 flex justify-around border-b border-[#333333]">
            <button
              onClick={() => setView('details')}
              className={cn(
                'relative border-b-2 px-4 py-2 text-sm font-medium transition-colors',
                view === 'details'
                  ? 'border-white text-white'
                  : 'border-transparent text-[#8b8b8b] hover:text-white',
              )}
            >
              Details
            </button>
            <button
              onClick={() => setView('exploit')}
              className={cn(
                'relative border-b-2 px-4 py-2 text-sm font-medium transition-colors',
                view === 'exploit'
                  ? 'border-white text-white'
                  : 'border-transparent text-[#8b8b8b] hover:text-white',
              )}
            >
              Exploit
            </button>
          </div>
        </DialogHeader>

        {/* Content */}
        <div className="overflow-y-auto bg-[#1a1a1a] p-6 pt-1">
          {view === 'details' && (
            <div className="space-y-6">
              {/* Flag Code */}
              <div>
                <h3 className="mb-3 text-sm font-medium text-gray-200">
                  Flag Code
                </h3>
                <div className="rounded border border-[#333333] bg-[#2a2a2a] p-4">
                  <div className="flex items-center justify-between">
                    <code className="font-mono text-sm break-all text-gray-300">
                      {flag.flag_code}
                    </code>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={handleCopyCode}
                      className="ml-2 h-8 px-2 text-[#8b8b8b] hover:bg-[#333333] hover:text-white"
                    >
                      <Copy className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
              </div>

              <div>
                <div className="mb-2 flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4 text-gray-200" />
                  <span className="text-sm font-medium text-gray-200">
                    Message
                  </span>
                </div>
                <div className="text-sm text-[#8b8b8b]">{flag.msg}</div>
              </div>

              {/* Details Grid - matching table layout */}
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-6">
                  <div>
                    <div className="mb-2 flex items-center gap-2">
                      <Server className="h-4 w-4 text-gray-200" />
                      <span className="text-base font-medium text-gray-200">
                        Service
                      </span>
                    </div>
                    <div className="text-[#8b8b8b]">
                      <div className="text-base">{flag.service_name}</div>
                      <div className="text-sm">Port {flag.port_service}</div>
                    </div>
                  </div>

                  <div>
                    <div className="mb-2 flex items-center gap-2">
                      <User className="h-4 w-4 text-gray-200" />
                      <span className="text-base font-medium text-gray-200">
                        Victim
                      </span>
                    </div>
                    <div className="text-sm text-[#8b8b8b]">
                      <div>
                        {' '}
                        Team Id{' '}
                        <span className="font-mono text-base text-[#8b8b8b]">
                          #{flag.team_id}
                        </span>
                      </div>
                      <div className="text-sm">
                        with{' '}
                        <span className="text-rose-400/80">
                          {flag.exploit_name}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-6">
                  <div>
                    <div className="mb-2 flex items-center gap-2">
                      <Clock className="h-4 w-4 text-gray-200" />
                      <span className="text-base font-medium text-gray-200">
                        Submitted
                      </span>
                    </div>
                    <div className="text-base text-[#8b8b8b]">
                      <div>
                        {new Date(flag.submit_time).toLocaleTimeString()}
                      </div>
                      <div className="text-xs">
                        {new Date(flag.submit_time).toDateString()}
                      </div>
                    </div>
                  </div>

                  <div>
                    <div className="mb-2 flex items-center gap-2">
                      <Hourglass className="h-4 w-4 text-gray-200" />
                      <span className="text-sm font-medium text-gray-200">
                        Duration
                      </span>
                    </div>
                    <div className="text-[#8b8b8b]">
                      {flag.response_time - flag.submit_time} seconds
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {view === 'exploit' && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Code2 className="h-4 w-4 text-[#8b8b8b]" />
                  <h3 className="text-sm font-medium text-white">
                    Exploit Code{' '}
                    <span className="font-mono font-light text-rose-500">
                      {flag.exploit_name}
                    </span>
                  </h3>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() =>
                    navigator.clipboard.writeText(
                      `#!/usr/bin/env\npython3\nfrom cookiefarm import exploit_manager\nimport requests\n\n@exploit_manager\ndef exploit(ip, port, name_service, flag_ids: list):\tflag = requests.get(f"http://{ip}:{port}/get-flag").text\n\tprint(flag)\n`,
                    )
                  }
                  className="h-8 px-2 text-[#8b8b8b] hover:bg-[#333333] hover:text-white"
                >
                  <Copy className="mr-1 h-3 w-3" />
                  Copy
                </Button>
              </div>

              <div className="rounded border border-[#333333] bg-[#2a2a2a]">
                <pre className="overflow-x-auto p-4 font-mono text-sm text-[#8b8b8b]">
                  <code>{`#!/usr/bin/env python3
from cookiefarm import exploit_manager
import requests

@exploit_manager
def exploit(ip, port, name_service, flag_ids: list):
  flag = requests.get(f"http://{ip}:{port}/get-flag").text
  print(flag)

`}</code>
                </pre>
              </div>
            </div>
          )}
        </div>

        {/* Delete Confirmation Dialog */}
        {showDeleteConfirm && (
          <div className="absolute inset-0 flex items-center justify-center bg-black/50">
            <div className="mx-4 max-w-md rounded-lg border border-[#333333] bg-[#2a2a2a] p-6">
              <div className="mb-4 flex items-center gap-3">
                <Trash2 className="h-5 w-5 text-red-400" />
                <h3 className="text-lg font-medium text-white">Delete Flag</h3>
              </div>
              <p className="mb-6 text-[#8b8b8b]">
                Are you sure you want to delete this flag? This action cannot be
                undone.
              </p>
              <div className="flex justify-end gap-3">
                <Button
                  variant="ghost"
                  onClick={() => setShowDeleteConfirm(false)}
                  className="text-[#8b8b8b] hover:bg-[#333333] hover:text-white"
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleDelete}
                  className="bg-red-400 text-white hover:bg-red-500"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Delete
                </Button>
              </div>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
