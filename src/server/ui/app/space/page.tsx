"use client";

import { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Loader2, Plus, RefreshCw } from "lucide-react";
import {
  getSpaces,
  createSpace,
  deleteSpace,
  getSpaceConfigs,
  updateSpaceConfigs,
} from "@/api/models/space";
import { Space } from "@/types";

export default function SpacesPage() {
  const t = useTranslations("space");
  const router = useRouter();

  const [spaces, setSpaces] = useState<Space[]>([]);
  const [selectedSpace, setSelectedSpace] = useState<Space | null>(null);
  const [isLoadingSpaces, setIsLoadingSpaces] = useState(true);
  const [isCreatingSpace, setIsCreatingSpace] = useState(false);
  const [isRefreshingSpaces, setIsRefreshingSpaces] = useState(false);
  const [spaceFilterText, setSpaceFilterText] = useState("");

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [spaceToDelete, setSpaceToDelete] = useState<Space | null>(null);
  const [isDeletingSpace, setIsDeletingSpace] = useState(false);

  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [createConfigValue, setCreateConfigValue] = useState("{}");

  const [configDialogOpen, setConfigDialogOpen] = useState(false);
  const [configEditValue, setConfigEditValue] = useState("");
  const [isSavingConfig, setIsSavingConfig] = useState(false);
  const [configEditTarget, setConfigEditTarget] = useState<Space | null>(null);

  const filteredSpaces = spaces.filter((space) =>
    space.id.toLowerCase().includes(spaceFilterText.toLowerCase())
  );

  const loadSpaces = async () => {
    try {
      setIsLoadingSpaces(true);
      const res = await getSpaces();
      if (res.code !== 0) {
        console.error(res.message);
        return;
      }
      setSpaces(res.data || []);
    } catch (error) {
      console.error("Failed to load spaces:", error);
    } finally {
      setIsLoadingSpaces(false);
    }
  };

  useEffect(() => {
    loadSpaces();
  }, []);

  const handleOpenCreateDialog = () => {
    setCreateConfigValue("{}");
    setCreateDialogOpen(true);
  };

  const handleCreateSpace = async () => {
    try {
      const configs = JSON.parse(createConfigValue);
      setIsCreatingSpace(true);
      const res = await createSpace(configs);
      if (res.code !== 0) {
        console.error(res.message);
        return;
      }
      await loadSpaces();
      setCreateDialogOpen(false);
    } catch (error) {
      console.error("Failed to create space:", error);
      alert(t("invalidJson"));
    } finally {
      setIsCreatingSpace(false);
    }
  };

  const handleDeleteSpace = async () => {
    if (!spaceToDelete) return;
    try {
      setIsDeletingSpace(true);
      const res = await deleteSpace(spaceToDelete.id);
      if (res.code !== 0) {
        console.error(res.message);
        return;
      }
      if (selectedSpace?.id === spaceToDelete.id) {
        setSelectedSpace(null);
      }
      await loadSpaces();
    } catch (error) {
      console.error("Failed to delete space:", error);
    } finally {
      setIsDeletingSpace(false);
      setDeleteDialogOpen(false);
      setSpaceToDelete(null);
    }
  };

  const handleRefreshSpaces = async () => {
    setIsRefreshingSpaces(true);
    await loadSpaces();
    setIsRefreshingSpaces(false);
  };

  const handleViewConfig = async (space: Space) => {
    try {
      setConfigEditTarget(space);
      let configs = space.configs;

      const res = await getSpaceConfigs(space.id);
      if (res.code === 0 && res.data) {
        configs = res.data.configs;
      }

      setConfigEditValue(JSON.stringify(configs, null, 2));
      setConfigDialogOpen(true);
    } catch (error) {
      console.error("Failed to load config:", error);
    }
  };

  const handleSaveConfig = async () => {
    if (!configEditTarget) return;

    try {
      const configs = JSON.parse(configEditValue);
      setIsSavingConfig(true);

      const res = await updateSpaceConfigs(configEditTarget.id, configs);
      if (res.code !== 0) {
        console.error(res.message);
        return;
      }
      await loadSpaces();
      setConfigDialogOpen(false);
    } catch (error) {
      console.error("Failed to save config:", error);
      alert(t("invalidJson"));
    } finally {
      setIsSavingConfig(false);
    }
  };

  const handleGoToPages = (spaceId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    router.push(`/space/${spaceId}/pages`);
  };

  return (
    <div className="h-full bg-background p-6">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">{t("spaceList")}</h1>
            <p className="text-sm text-muted-foreground mt-1">
              {t("spaceListDescription") || "管理所有 Space"}
            </p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              onClick={handleOpenCreateDialog}
            >
              <Plus className="h-4 w-4 mr-2" />
              {t("createSpace")}
            </Button>
            <Button
              variant="outline"
              onClick={handleRefreshSpaces}
              disabled={isRefreshingSpaces}
            >
              {isRefreshingSpaces ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  {t("loading")}
                </>
              ) : (
                <>
                  <RefreshCw className="h-4 w-4 mr-2" />
                  {t("refresh")}
                </>
              )}
            </Button>
          </div>
        </div>

        <Input
          type="text"
          placeholder={t("filterById")}
          value={spaceFilterText}
          onChange={(e) => setSpaceFilterText(e.target.value)}
          className="max-w-sm"
        />

        <div className="rounded-md border">
          {isLoadingSpaces ? (
            <div className="flex items-center justify-center h-64">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : filteredSpaces.length === 0 ? (
            <div className="flex items-center justify-center h-64">
              <p className="text-sm text-muted-foreground">
                {spaces.length === 0 ? t("noData") : t("noMatching")}
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("spaceId")}</TableHead>
                  <TableHead>{t("createdAt")}</TableHead>
                  <TableHead>{t("actions")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredSpaces.map((space) => (
                  <TableRow
                    key={space.id}
                    className="cursor-pointer"
                    data-state={selectedSpace?.id === space.id ? "selected" : undefined}
                    onClick={() => setSelectedSpace(space)}
                  >
                    <TableCell className="font-mono">
                      {space.id}
                    </TableCell>
                    <TableCell>
                      {new Date(space.created_at).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={(e) => handleGoToPages(space.id, e)}
                        >
                          {t("pages") || "Pages"}
                        </Button>
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleViewConfig(space);
                          }}
                        >
                          {t("config")}
                        </Button>
                        <Button
                          variant="secondary"
                          size="sm"
                          className="text-destructive hover:text-destructive"
                          onClick={(e) => {
                            e.stopPropagation();
                            setSpaceToDelete(space);
                            setDeleteDialogOpen(true);
                          }}
                        >
                          {t("delete")}
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </div>
      </div>

      {/* Delete Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("deleteConfirmTitle")}</AlertDialogTitle>
            <AlertDialogDescription>
              {t("deleteSpaceConfirm")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeletingSpace}>
              {t("cancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteSpace}
              disabled={isDeletingSpace}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeletingSpace ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  {t("deleting")}
                </>
              ) : (
                t("delete")
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Config Dialog */}
      <AlertDialog open={configDialogOpen} onOpenChange={setConfigDialogOpen}>
        <AlertDialogContent className="max-w-2xl">
          <AlertDialogHeader>
            <AlertDialogTitle>{t("editConfigsTitle")}</AlertDialogTitle>
          </AlertDialogHeader>
          <div className="py-4">
            <textarea
              className="w-full h-64 p-2 font-mono text-sm border rounded-md"
              value={configEditValue}
              onChange={(e) => setConfigEditValue(e.target.value)}
              placeholder={t("configsPlaceholder")}
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isSavingConfig}>
              {t("cancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleSaveConfig}
              disabled={isSavingConfig}
            >
              {isSavingConfig ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  {t("saving")}
                </>
              ) : (
                t("save")
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Create Space Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{t("createSpaceTitle")}</DialogTitle>
            <DialogDescription>{t("createSpaceDescription")}</DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <textarea
              className="w-full h-64 p-2 font-mono text-sm border rounded-md"
              value={createConfigValue}
              onChange={(e) => setCreateConfigValue(e.target.value)}
              placeholder={t("configsPlaceholder")}
            />
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setCreateDialogOpen(false)}
              disabled={isCreatingSpace}
            >
              {t("cancel")}
            </Button>
            <Button
              onClick={handleCreateSpace}
              disabled={isCreatingSpace}
            >
              {isCreatingSpace ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  {t("creating")}
                </>
              ) : (
                t("create")
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
