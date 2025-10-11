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

function truncateMiddle(str: string, maxLength: number = 30): string {
  if (str.length <= maxLength) return str;

  const ellipsis = "...";
  const charsToShow = maxLength - ellipsis.length;
  const frontChars = Math.ceil(charsToShow / 2);
  const backChars = Math.floor(charsToShow / 2);

  return str.substring(0, frontChars) + ellipsis + str.substring(str.length - backChars);
}

function Node({ node, style, dragHandle, loadingNodes }: NodeProps) {
  const indent = node.level * 12;
  const isFolder = node.data.type === "folder";
  const isLoading = loadingNodes.has(node.id);
  const textRef = useRef<HTMLSpanElement>(null);
  const [displayName, setDisplayName] = useState(node.data.name);

  useEffect(() => {
    const updateDisplayName = () => {
      if (!textRef.current) return;

      const container = textRef.current.parentElement;
      if (!container) return;

      // Get available width (container width - icon width - gap - padding)
      const containerWidth = container.clientWidth;
      const iconWidth = isFolder ? 56 : 40; // Total width of icon and spacing
      const availableWidth = containerWidth - iconWidth;

      // Create temporary element to measure text width
      const tempSpan = document.createElement("span");
      tempSpan.style.visibility = "hidden";
      tempSpan.style.position = "absolute";
      tempSpan.style.fontSize = "14px"; // text-sm
      tempSpan.style.fontFamily = getComputedStyle(textRef.current).fontFamily;
      tempSpan.textContent = node.data.name;
      document.body.appendChild(tempSpan);

      const fullWidth = tempSpan.offsetWidth;
      document.body.removeChild(tempSpan);

      // If text width is less than available width, display full name
      if (fullWidth <= availableWidth) {
        setDisplayName(node.data.name);
        return;
      }

      // Calculate number of characters to display
      const charWidth = fullWidth / node.data.name.length;
      const maxChars = Math.floor(availableWidth / charWidth);

      setDisplayName(truncateMiddle(node.data.name, Math.max(10, maxChars)));
    };

    updateDisplayName();

    // Add window resize listener
    const resizeObserver = new ResizeObserver(updateDisplayName);
    if (textRef.current?.parentElement) {
      resizeObserver.observe(textRef.current.parentElement);
    }

    return () => {
      resizeObserver.disconnect();
    };
  }, [node.data.name, indent, isFolder]);

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
        className="flex items-center gap-1.5 flex-1 min-w-0"
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
        <span
          ref={textRef}
          className="min-w-0"
          title={node.data.name}
        >
          {displayName}
        </span>
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

  // Artifact related states
  const [artifacts, setArtifacts] = useState<Artifact[]>([]);
  const [selectedArtifact, setSelectedArtifact] = useState<Artifact | null>(
    null
  );
  const [isLoadingArtifacts, setIsLoadingArtifacts] = useState(true);

  // Load artifact list when component mounts
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

  // Load root directory files when artifact is selected
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

    // Return if already loaded
    if (node.data.isLoaded) return;

    // Mark as loading
    setLoadingNodes((prev) => new Set(prev).add(nodeId));

    try {
      // Load children using unified interface with artifact_id and path
      const children = await getListFiles(selectedArtifact.id, node.data.path);
      if (children.code !== 0 || !children.data) {
        console.error(children.message);
        return;
      }
      const files = formatFiles(node.data.path, children.data);

      // Update node data
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
      // Remove loading state
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

            {/* Artifact selector */}
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
                      <span className="mr-2 min-w-0 truncate" title={selectedArtifact.id}>
                        {selectedArtifact.id}
                      </span>
                      <ChevronDown className="h-4 w-4 opacity-50 shrink-0" />
                    </>
                  ) : (
                    <>
                      <span className="text-muted-foreground">
                        Select an artifact
                      </span>
                      <ChevronDown className="h-4 w-4 opacity-50 shrink-0" />
                    </>
                  )}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-[var(--radix-dropdown-menu-trigger-width)]">
                {artifacts.map((artifact) => (
                  <DropdownMenuItem
                    key={artifact.id}
                    onClick={() => handleArtifactSelect(artifact)}
                    title={artifact.id}
                  >
                    <span className="truncate block w-full">
                      {artifact.id}
                    </span>
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
