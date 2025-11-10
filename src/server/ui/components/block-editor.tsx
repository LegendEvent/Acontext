"use client";

import { useState, useEffect } from "react";
import * as React from "react";
import { useForm, type Resolver } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Plus, Loader2 } from "lucide-react";

// Block type configuration - easily extensible
export interface BlockTypeConfig {
  type: string;
  label: string;
  description: string;
  fields: {
    name: string;
    label: string;
    type: "text" | "textarea";
    placeholder?: string;
    required?: boolean;
  }[];
}

// Default block type configurations
export const DEFAULT_BLOCK_CONFIGS: BlockTypeConfig[] = [
  {
    type: "text",
    label: "Text Block",
    description: "Add text notes",
    fields: [
      {
        name: "title",
        label: "Use When",
        type: "text",
        placeholder: "When should this be used?",
        required: true,
      },
      {
        name: "notes",
        label: "Notes",
        type: "textarea",
        placeholder: "Enter your notes...",
      },
    ],
  },
  {
    type: "sop",
    label: "SOP Block",
    description: "Add standard operating procedure",
    fields: [
      {
        name: "title",
        label: "Use When",
        type: "text",
        placeholder: "When should this be used?",
        required: true,
      },
      {
        name: "preferences",
        label: "Preferences",
        type: "textarea",
        placeholder: "Enter any preferences or guidelines...",
      },
    ],
  },
];

export interface BlockEditorProps {
  mode: "create" | "edit";
  blockType: string;
  initialValues?: Record<string, string>;
  blockConfigs?: BlockTypeConfig[];
  onSave: (values: Record<string, string>) => Promise<void>;
  onCancel: () => void;
  open: boolean;
  isLoading?: boolean;
  translations?: {
    cancel?: string;
    save?: string;
    saving?: string;
    create?: string;
    creating?: string;
  };
}

export function BlockEditor({
  mode,
  blockType,
  initialValues = {},
  blockConfigs = DEFAULT_BLOCK_CONFIGS,
  onSave,
  onCancel,
  open,
  isLoading = false,
  translations = {},
}: BlockEditorProps) {
  const config = blockConfigs.find((c) => c.type === blockType);

  // Build Zod schema dynamically based on field configuration
  const schema = React.useMemo(() => {
    if (!config) return z.object({});

    const schemaFields: Record<string, z.ZodString> = {};
    config.fields.forEach((field) => {
      let fieldSchema = z.string();
      if (field.required) {
        fieldSchema = fieldSchema.min(1, `${field.label} is required`);
      }
      schemaFields[field.name] = fieldSchema;
    });

    return z.object(schemaFields);
  }, [config]);

  type FormValues = Record<string, string>;

  // Initialize react-hook-form
  const form = useForm<FormValues>({
    resolver: zodResolver(schema) as unknown as Resolver<FormValues>,
    defaultValues: initialValues,
  });

  // Update form values when initialValues or open changes
  useEffect(() => {
    if (open && config) {
      const values: Record<string, string> = {};
      config.fields.forEach((field) => {
        values[field.name] = initialValues[field.name] || "";
      });
      form.reset(values);
    }
  }, [open, initialValues, config, form, schema]);

  const handleSave = async (values: FormValues) => {
    await onSave(values);
  };

  const handleCancel = () => {
    onCancel();
    form.reset();
  };

  const t = {
    cancel: translations.cancel || "Cancel",
    save: translations.save || "Save",
    saving: translations.saving || "Saving...",
    create: translations.create || "Create",
    creating: translations.creating || "Creating...",
  };

  if (!config) return null;

  const title = mode === "create" ? `Create ${config.label}` : `Edit ${config.label}`;
  const description = config.description;

  return (
    <AlertDialog open={open} onOpenChange={handleCancel}>
      <AlertDialogContent className="max-w-2xl">
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{description}</AlertDialogDescription>
        </AlertDialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSave)} className="space-y-4">
            <div className="py-4 space-y-4">
              {config?.fields.map((field) => (
                <FormField
                  key={field.name}
                  control={form.control}
                  name={field.name}
                  render={({ field: formField }) => (
                    <FormItem>
                      <FormLabel>
                        {field.label}
                        {field.required && <span className="text-destructive ml-1">*</span>}
                      </FormLabel>
                      <FormControl>
                        {field.type === "text" ? (
                          <Input
                            type="text"
                            placeholder={field.placeholder}
                            {...formField}
                          />
                        ) : (
                          <Textarea
                            className="min-h-[120px]"
                            placeholder={field.placeholder}
                            {...formField}
                          />
                        )}
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              ))}
            </div>
            <AlertDialogFooter>
              <AlertDialogCancel disabled={isLoading} onClick={handleCancel}>
                {t.cancel}
              </AlertDialogCancel>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    {mode === "create" ? t.creating : t.saving}
                  </>
                ) : (
                  <>
                    {mode === "create" && <Plus className="h-4 w-4" />}
                    {mode === "create" ? t.create : t.save}
                  </>
                )}
              </Button>
            </AlertDialogFooter>
          </form>
        </Form>
      </AlertDialogContent>
    </AlertDialog>
  );
}

// Convenience hooks for using the editor
export function useBlockEditor() {
  const [isOpen, setIsOpen] = useState(false);
  const [mode, setMode] = useState<"create" | "edit">("create");
  const [blockType, setBlockType] = useState<string>("text");
  const [initialValues, setInitialValues] = useState<Record<string, string>>({});

  const openCreate = (type: string) => {
    setMode("create");
    setBlockType(type);
    setInitialValues({});
    setIsOpen(true);
  };

  const openEdit = (type: string, values: Record<string, string>) => {
    setMode("edit");
    setBlockType(type);
    setInitialValues(values);
    setIsOpen(true);
  };

  const close = () => {
    setIsOpen(false);
  };

  return {
    isOpen,
    mode,
    blockType,
    initialValues,
    openCreate,
    openEdit,
    close,
  };
}

