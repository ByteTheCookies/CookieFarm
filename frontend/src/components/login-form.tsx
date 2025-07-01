"use client";

import axios from "axios";
import { cn } from "@/lib/utils";
import Image from "next/image";
import { useState } from "react";
import { Eye, EyeOff } from "lucide-react";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage as OriginalFormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";

const loginSchema = z.object({
  password: z.string().min(1, "Password is required"),
});

function FormMessage(props: React.ComponentProps<typeof OriginalFormMessage>) {
  const { children, ...rest } = props;

  return (
    <OriginalFormMessage
      {...rest}
      className={cn("text-red-500 text-sm font-medium", props.className)}
    >
      {children}
    </OriginalFormMessage>
  );
}


export function LoginForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const form = useForm({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      password: "",
    },
  });

  function onSubmit(values: z.infer<typeof loginSchema>) {
    const formData = new FormData();
    formData.append("password", values.password);

    axios
      .post("/api/login", formData)
      .then((response) => {
        console.log("Login successful:", response.data);
      })
      .catch((error) => {
        console.error("Login failed:", error);
      });
  }


  const [showPassword, setShowPassword] = useState(false);

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
        <div
          className={cn(
            "flex flex-col gap-6 p-9 rounded-xl shadow-2xl bg-black/30 backdrop-blur-xl border border-gray-800/50",

            className,
          )}
          {...props}
        >
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)}>
              <div className="flex flex-col gap-6">
                <div className="flex flex-col items-center gap-2 mb-7">
                  <a
                    href="#"
                    className="flex flex-col items-center gap-2 font-medium text-white"
                  >
                    <div className="flex size-15 items-center justify-center rounded-md">
                      <Image
                        src="images/logo.png"
                        alt="logo"
                        width={200}
                        height={200}
                      />
                    </div>
                    <span className="sr-only">CookieFarm</span>
                  </a>
                  <h1 className="text-3xl text-nowrap font-bold text-white">
                    Welcome to CookieFarm
                  </h1>

                </div>

                <FormField
                  control={form.control}
                  name="password"
                  render={({ field, fieldState }) => (
                    <FormItem>
                      <FormLabel
                        className={fieldState.error ? "!text-red-500" : ""}
                      >
                        Password
                      </FormLabel>
                      <FormControl>
                        {/* Wrap Input and Button in a single parent to ensure only one child */}
                        <div className="relative">
                          <Input
                            placeholder="3spositOne"
                            type={showPassword ? "text" : "password"}
                            {...field}
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
                      </FormControl>
                      <FormDescription>
                        Enter your password to access to the server.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <Button type="submit" className="w-full text-white">
                  Login
                </Button>
              </div>
              {/* </div> */}
            </form>
          </Form>
        </div>
      </div>
    </div>
  );
}
