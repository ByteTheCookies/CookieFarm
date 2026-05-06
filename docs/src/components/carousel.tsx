"use client";
// ─── Carousel Data ───────────────────────────────────────────────────────────

import { useCallback, useEffect, useMemo, useState } from "react";
import Image from "next/image";
import { ChevronLeft, ChevronRight } from "lucide-react";

const carouselSlides = [
  {
    id: "dashboard",
    image: "/carousel/dashboard.png",
    title: "Live Dashboard",
    description: "Monitor your exploits in real-time with a clean, responsive interface",
  },
  {
    id: "flags",
    image: "/carousel/flags.png",
    title: "Flag Submission",
    description: "Automatic deduplication and submission to the scoreboard every tick",
  },
  {
    id: "charts",
    image: "/carousel/charts.png",
    title: "Charts & Analytics",
    description: "Visualize your performance with detailed charts and analytics",
  },
  {
    id: "config",
    image: "/carousel/config.png",
    title: "Easy Configuration",
    description: "Configure your in the dashboard and let CookieFarm handle the rest",
  },
];

// ─── Carousel ────────────────────────────────────────────────────────────────

export function Carousel() {
  const [currentSlide, setCurrentSlide] = useState(0);
  const [isHovered, setIsHovered] = useState(false);
  const [direction, setDirection] = useState<"left" | "right">("right");

  const nextSlide = useCallback(() => {
    setDirection("right");
    setCurrentSlide((prev) => (prev + 1) % carouselSlides.length);
  }, []);

  const prevSlide = useCallback(() => {
    setDirection("left");
    setCurrentSlide((prev) => (prev - 1 + carouselSlides.length) % carouselSlides.length);
  }, []);

  const goToSlide = (index: number) => {
    setDirection(index > currentSlide ? "right" : "left");
    setCurrentSlide(index);
  };

  const slideClassNameByIndex = useMemo(() => {
    return (index: number) => {
      const isCurrent = index === currentSlide;
      const isPrev =
        index < currentSlide ||
        (currentSlide === 0 && index === carouselSlides.length - 1 && direction === "left");

      if (isCurrent) return "translate-x-0 opacity-100";
      if (isPrev) return "-translate-x-full opacity-0";
      return "translate-x-full opacity-0";
    };
  }, [currentSlide, direction]);

  const contentClassNameByIndex = useMemo(() => {
    return (index: number) => {
      const isCurrent = index === currentSlide;
      return `transform transition-all duration-500 delay-200 ${isCurrent ? "translate-y-0 opacity-100" : "translate-y-4 opacity-0"}`;
    };
  }, [currentSlide]);

  // Auto-play
  useEffect(() => {
    if (isHovered) return;
    const interval = setInterval(nextSlide, 4000);
    return () => clearInterval(interval);
  }, [isHovered, nextSlide]);

  return (
    <div
      className="group relative w-full overflow-hidden rounded-lg border border-(--surface-border) bg-(--surface)"
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* Slides Container */}
      <div className="relative aspect-video w-full overflow-hidden">
        {carouselSlides.map((slide, index) => (
          <div
            key={slide.id}
            className={`absolute inset-0 transition-all duration-700 ease-out ${slideClassNameByIndex(index)}`}
          >
            <Image src={slide.image} alt={slide.title} fill className="object-cover" priority={index === 0} />
            {/* Gradient overlay */}
            <div className="absolute inset-0 bg-linear-to-t from-(--surface) via-transparent to-transparent" />

            {/* Content overlay */}
            <div className="absolute bottom-0 left-0 right-0 p-6">
              <div className={contentClassNameByIndex(index)}>
                <h3 className="mb-1 font-mono text-lg font-semibold text-foreground">{slide.title}</h3>
                <p className="text-sm text-muted-foreground">{slide.description}</p>
              </div>
            </div>
          </div>
        ))}

        {/* Navigation Arrows */}
        <button
          type="button"
          onClick={prevSlide}
          className="absolute left-3 top-1/2 -translate-y-1/2 flex h-10 w-10 items-center justify-center rounded-full border border-(--surface-border) bg-(--surface)/90 text-foreground opacity-0 backdrop-blur-sm transition-all duration-300 hover:border-(--green)/50 hover:text-(--green) group-hover:opacity-100"
          aria-label="Previous slide"
        >
          <ChevronLeft size={20} />
        </button>
        <button
          type="button"
          onClick={nextSlide}
          className="absolute right-3 top-1/2 -translate-y-1/2 flex h-10 w-10 items-center justify-center rounded-full border border-(--surface-border) bg-(--surface)/90 text-foreground opacity-0 backdrop-blur-sm transition-all duration-300 hover:border-(--green)/50 hover:text-(--green) group-hover:opacity-100"
          aria-label="Next slide"
        >
          <ChevronRight size={20} />
        </button>
      </div>

      {/* Dots Indicator */}
      <div className="flex items-center justify-center gap-2 py-4">
        {carouselSlides.map((slide, index) => (
          <button
            type="button"
            key={slide.id}
            onClick={() => goToSlide(index)}
            className={`h-2 rounded-full transition-all duration-300 ${index === currentSlide ? "w-6 bg-(--green)" : "w-2 bg-(--surface-border) hover:bg-muted-foreground"}`}
            aria-label={`Go to slide ${index + 1}`}
          />
        ))}
      </div>

      {/* Progress bar */}
      <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-(--surface-border)">
        <div
          className="h-full bg-(--green) transition-all duration-300 ease-linear"
          style={{
            width: `${((currentSlide + 1) / carouselSlides.length) * 100}%`,
          }}
        />
      </div>
    </div>
  );
}
