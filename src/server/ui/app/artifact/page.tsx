"use client";

import { useState, useRef, useEffect } from "react";
import { Tree, NodeRendererProps, TreeApi, NodeApi } from "react-arborist";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import {
  ChevronRight,
  File,
  Folder,
  FolderOpen,
  Loader2,
  ChevronDown,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { getArtifacts, getListFiles } from "@/api/models/artifact";
import { Artifact, ListFilesResp } from "@/types";

interface TreeNode {
  id: string;
  name: string;
  type: "folder" | "file";
  path: string;
  children?: TreeNode[];
  isLoaded?: boolean;
}

interface NodeProps extends NodeRendererProps<TreeNode> {
  loadingNodes: Set<string>;
}

function Node({ node, style, dragHandle, loadingNodes }: NodeProps) {
  const indent = node.level * 12;
  const isFolder = node.data.type === "folder";
  const isLoading = loadingNodes.has(node.id);

  return (
    <div
      ref={dragHandle}
      style={style}
      className={cn(
        "flex items-center cursor-pointer px-2 py-1.5 text-sm rounded-md transition-colors",
        "hover:bg-accent hover:text-accent-foreground",
        node.isSelected && "bg-accent text-accent-foreground",
        node.state.isDragging && "opacity-50"
      )}
      onClick={() => {
        if (isFolder) {
          node.toggle();
        } else {
          node.select();
        }
      }}
    >
      <div
        style={{ marginLeft: `${indent}px` }}
        className="flex items-center gap-1.5 flex-1"
      >
        {isFolder ? (
          <>
            {isLoading ? (
              <Loader2 className="h-4 w-4 shrink-0 animate-spin text-muted-foreground" />
            ) : (
              <ChevronRight
                className={cn(
                  "h-4 w-4 shrink-0 transition-transform duration-200",
                  node.isOpen && "rotate-90"
                )}
              />
            )}
            {node.isOpen ? (
              <FolderOpen className="h-4 w-4 shrink-0 text-muted-foreground" />
            ) : (
              <Folder className="h-4 w-4 shrink-0 text-muted-foreground" />
            )}
          </>
        ) : (
          <>
            <span className="w-4" />
            <File className="h-4 w-4 shrink-0 text-muted-foreground" />
          </>
        )}
        <span className="truncate">{node.data.name}</span>
      </div>
    </div>
  );
}

export default function ArtifactPage() {
  const treeRef = useRef<TreeApi<TreeNode>>(null);
  const [selectedFile, setSelectedFile] = useState<TreeNode | null>(null);
  const [loadingNodes, setLoadingNodes] = useState<Set<string>>(new Set());
  const [treeData, setTreeData] = useState<TreeNode[]>([]);
  const [isInitialLoading, setIsInitialLoading] = useState(false);

  // Artifact 相关状态
  const [artifacts, setArtifacts] = useState<Artifact[]>([]);
  const [selectedArtifact, setSelectedArtifact] = useState<Artifact | null>(
    null
  );
  const [isLoadingArtifacts, setIsLoadingArtifacts] = useState(true);

  // 组件挂载时加载 artifact 列表
  useEffect(() => {
    const loadArtifacts = async () => {
      try {
        setIsLoadingArtifacts(true);
        const res = await getArtifacts();
        if (res.code !== 0) {
          console.error(res.message);
          return;
        }
        setArtifacts(res.data || []);
      } catch (error) {
        console.error("Failed to load artifacts:", error);
      } finally {
        setIsLoadingArtifacts(false);
      }
    };

    loadArtifacts();
  }, []);

  const formatFiles = (path: string, res: ListFilesResp) => {
    const files: TreeNode[] = res.files.map((file) => ({
      id: file.path,
      name: file.filename,
      type: "file",
      path: file.path,
      isLoaded: false,
    }));
    const directories: TreeNode[] = res.directories.map((directory) => ({
      id: directory,
      name: directory,
      type: "folder",
      path: `${path}${directory}/`,
      isLoaded: false,
    }));
    return [...directories, ...files];
  };

  // 当选择 artifact 时，加载根目录文件
  const handleArtifactSelect = async (artifact: Artifact) => {
    setSelectedArtifact(artifact);
    setTreeData([]);
    setSelectedFile(null);

    try {
      setIsInitialLoading(true);
      const res = await getListFiles(artifact.id, "/");
      if (res.code !== 0 || !res.data) {
        console.error(res.message);
        return;
      }
      setTreeData(formatFiles("/", res.data));
    } catch (error) {
      console.error("Failed to load files:", error);
    } finally {
      setIsInitialLoading(false);
    }
  };

  const handleToggle = async (nodeId: string) => {
    const node = treeRef.current?.get(nodeId);
    if (!node || node.data.type !== "folder" || !selectedArtifact) return;

    // 如果已经加载过，直接返回
    if (node.data.isLoaded) return;

    // 标记为加载中
    setLoadingNodes((prev) => new Set(prev).add(nodeId));

    try {
      // 使用统一的接口加载子项，传入 artifact_id 和 path
      const children = await getListFiles(selectedArtifact.id, node.data.path);
      if (children.code !== 0 || !children.data) {
        console.error(children.message);
        return;
      }
      const files = formatFiles(node.data.path, children.data);

      // 更新节点数据
      setTreeData((prevData) => {
        const updateNode = (nodes: TreeNode[]): TreeNode[] => {
          return nodes.map((n) => {
            if (n.id === nodeId) {
              return {
                ...n,
                children: files,
                isLoaded: true,
              };
            }
            if (n.children) {
              return {
                ...n,
                children: updateNode(files),
              };
            }
            return n;
          });
        };
        return updateNode(prevData);
      });
    } catch (error) {
      console.error("Failed to load children:", error);
    } finally {
      // 移除加载状态
      setLoadingNodes((prev) => {
        const next = new Set(prev);
        next.delete(nodeId);
        return next;
      });
    }
  };

  const handleSelect = (nodes: NodeApi<TreeNode>[]) => {
    const node = nodes[0];
    if (node && node.data.type === "file") {
      setSelectedFile(node.data);
    }
  };

  return (
    <ResizablePanelGroup direction="horizontal" className="h-screen">
      <ResizablePanel defaultSize={30} minSize={20} maxSize={40}>
        <div className="h-full bg-background p-4">
          <div className="mb-4 space-y-3">
            <h2 className="text-lg font-semibold">File Explorer</h2>

            {/* Artifact 选择器 */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="outline"
                  className="w-full justify-between"
                  disabled={isLoadingArtifacts}
                >
                  {isLoadingArtifacts ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin" />
                      <span className="ml-2">Loading...</span>
                    </>
                  ) : selectedArtifact ? (
                    <>
                      <span>{selectedArtifact.id}</span>
                      <ChevronDown className="h-4 w-4 opacity-50" />
                    </>
                  ) : (
                    <>
                      <span className="text-muted-foreground">
                        Select an artifact
                      </span>
                      <ChevronDown className="h-4 w-4 opacity-50" />
                    </>
                  )}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-[var(--radix-dropdown-menu-trigger-width)]">
                {artifacts.map((artifact) => (
                  <DropdownMenuItem
                    key={artifact.id}
                    onClick={() => handleArtifactSelect(artifact)}
                  >
                    {artifact.id}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          <div className="h-[calc(100vh-11rem)]">
            {!selectedArtifact ? (
              <div className="flex items-center justify-center h-full">
                <p className="text-sm text-muted-foreground">
                  Select an artifact to view files
                </p>
              </div>
            ) : isInitialLoading ? (
              <div className="flex items-center justify-center h-full">
                <div className="flex flex-col items-center gap-2">
                  <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                  <p className="text-sm text-muted-foreground">
                    Loading files...
                  </p>
                </div>
              </div>
            ) : (
              <Tree
                ref={treeRef}
                data={treeData}
                openByDefault={false}
                width="100%"
                height={800}
                indent={12}
                rowHeight={32}
                className="p-2"
                onToggle={handleToggle}
                onSelect={handleSelect}
              >
                {(props) => <Node {...props} loadingNodes={loadingNodes} />}
              </Tree>
            )}
          </div>
        </div>
      </ResizablePanel>
      <ResizableHandle withHandle />
      <ResizablePanel>
        <div className="h-full bg-background p-4">
          <h2 className="mb-4 text-lg font-semibold">Content</h2>
          <div className="rounded-md border bg-card p-4">
            {selectedFile ? (
              <div>
                <h3 className="text-base font-medium mb-3">
                  {selectedFile.name}
                </h3>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">
                Select a file from the tree to view its content
              </p>
            )}
          </div>
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>
  );
}
