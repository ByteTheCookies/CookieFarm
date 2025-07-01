'use client';

import * as React from 'react';
import { Check, ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

interface StepProps {
  title: string;
  description?: string;
  isCompleted?: boolean;
  isActive?: boolean;
}

const Step: React.FC<StepProps> = ({
  title,
  description,
  isCompleted,
  isActive,
}) => {
  return (
    <div className="flex items-center">
      <div className="relative flex items-center justify-center">
        <div
          className={cn(
            'flex h-8 w-8 items-center justify-center rounded-full border-2',
            isCompleted
              ? 'border-primary bg-primary text-primary-foreground'
              : isActive
                ? 'border-primary'
                : 'border-muted',
          )}
        >
          {isCompleted ? (
            <Check className="h-4 w-4" />
          ) : (
            <span className="text-sm font-medium">{title[0]}</span>
          )}
        </div>
      </div>
      <div className="ml-4">
        <p
          className={cn(
            'text-sm font-medium',
            isActive || isCompleted
              ? 'text-foreground'
              : 'text-muted-foreground',
          )}
        >
          {title}
        </p>
        {description && (
          <p className="text-muted-foreground text-sm">{description}</p>
        )}
      </div>
    </div>
  );
};

interface StepperProps {
  steps: Array<{ title: string; description?: string }>;
  currentStep: number;
  onStepChange: (step: number) => void;
}

export function Stepper({ steps, currentStep, onStepChange }: StepperProps) {
  return (
    <div className="mx-auto w-full max-w-3xl">
      <div className="mb-8 flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
        {steps.map((step, index) => (
          <React.Fragment key={step.title}>
            <Step
              title={step.title}
              description={step.description}
              isCompleted={index < currentStep}
              isActive={index === currentStep}
            />
            {index < steps.length - 1 && (
              <ChevronRight className="text-muted-foreground hidden md:block" />
            )}
          </React.Fragment>
        ))}
      </div>
      <div className="flex justify-between">
        <Button
          variant="outline"
          onClick={() => onStepChange(currentStep - 1)}
          disabled={currentStep === 0}
        >
          Previous
        </Button>
        <Button
          onClick={() => onStepChange(currentStep + 1)}
          disabled={currentStep === steps.length - 1}
        >
          {currentStep === steps.length - 1 ? 'Finish' : 'Next'}
        </Button>
      </div>
    </div>
  );
}
