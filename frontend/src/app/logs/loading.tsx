export default function Loading() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="flex flex-col items-center space-y-4">
        <div className="relative">
          <div className="h-12 w-12 animate-spin rounded-full border-4 border-gray-200 border-t-blue-600"></div>
        </div>
        <p className="text-muted-foreground animate-pulse text-sm">
          Loading...
        </p>
      </div>
    </div>
  );
}
