"use client"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import Image from "next/image"
import { useState } from "react"
import { Eye, EyeOff } from "lucide-react"

export function LoginForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const [showPassword, setShowPassword] = useState(false)

  return (
    <div className="relative min-h-[90vh] w-full">
      {/* Background image covering the entire page */}
      <div className="fixed inset-0 z-0">
        <Image
          src="images/background-login.png"
          alt="background"
          fill
          className="object-cover brightness-70"
        />
      </div>

      {/* Login form container */}
      <div className="relative z-10 min-h-[90vh] flex items-center justify-center p-4">
        <div className={cn("flex flex-col gap-6 p-9 rounded-xl shadow-2xl bg-black/30 backdrop-blur-xl border border-gray-800/50", className)} {...props}>
          <form>
            <div className="flex flex-col gap-6">
              <div className="flex flex-col items-center gap-2 mb-7">
                <a
                  href="#"
                  className="flex flex-col items-center gap-2 font-medium text-white"
                >
                  <div className="flex size-15 items-center justify-center rounded-md">
                    <Image src="images/logo.png" alt="logo" width={200} height={200} />
                  </div>
                  <span className="sr-only">CookieFarm</span>
                </a>
                <h1 className="text-3xl text-nowrap font-bold text-white">Welcome to CookieFarm</h1>
              </div>
              <div className="flex flex-col gap-6">
                <div className="grid gap-3">
                  <Label htmlFor="password" className="text-white">Password</Label>
                  <div className="relative">
                    <Input
                      id="password"
                      type={showPassword ? "text" : "password"}
                      placeholder="3Sposit0n3"
                      required
                      className="bg-gray-900/50 border-gray-700 text-white placeholder:text-gray-400"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent text-white"
                      onClick={() => setShowPassword(!showPassword)}
                    >
                      {showPassword ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>
                <Button type="submit" className="w-full text-white">
                  Login
                </Button>
              </div>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
