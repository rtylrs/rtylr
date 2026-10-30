import React, { useEffect, useRef, useState } from "react";
import {
  Archive,
  BookOpen,
  BriefcaseBusiness,
  Check,
  Code2,
  Coffee,
  Folder,
  Globe2,
  Heart,
  Leaf,
  Lightbulb,
  Music2,
  Pencil,
  Plane,
  Plus,
  ShoppingBag,
  X,
} from "lucide-react";
import type { SaveProjectRequest } from "@voxire/voxie/hub/memory/v1/types";
import { useFocusTrap } from "../../hooks/useFocusTrap";
import type { CaretProject } from "../../types/chat";
import { caretRadius, caretTokens } from "../../ui/tokens";

export interface ProjectLabels {
  heading: string;
  allChats: string;
  create: string;
  createTitle: string;
  editTitle: string;
  name: string;
  context: string;
  contextHint: string;
  icon: string;
  color: string;
  save: string;
  cancel: string;
  archive: string;
  archiveTitle: string;
  archiveMessage: string;
  confirmArchive: string;
  empty: string;
  loadFailed: string;
  required: string;
  loading: string;
  saving: string;
  saveFailed: string;
  archiveFailed: string;
}

interface ProjectSidebarProps {
  projects: CaretProject[];
  activeId: string | null | undefined;
  labels?: Partial<ProjectLabels>;
  isLoading?: boolean;
  loadError?: boolean;
  onSelect: (id: string | null) => void;
  onSave: (project: SaveProjectRequest) => Promise<CaretProject>;
  onArchive: (id: string) => Promise<void>;
}

const ICONS = [
  ["folder", Folder],
  ["briefcase", BriefcaseBusiness],
  ["shopping-bag", ShoppingBag],
  ["heart", Heart],
  ["book", BookOpen],
  ["coffee", Coffee],
  ["globe", Globe2],
  ["code", Code2],
  ["lightbulb", Lightbulb],
  ["plane", Plane],
  ["leaf", Leaf],
  ["music", Music2],
] as const;
const COLORS = [
  ["neutral", "#74777d"],
  ["red", "#ef4444"],
  ["orange", "#f97316"],
  ["amber", "#eab308"],
  ["green", "#16a34a"],
  ["blue", "#1687f8"],
  ["violet", "#8b5cf6"],
  ["pink", "#ec4899"],
] as const;
const DEFAULT_LABELS: ProjectLabels = {
  heading: "Projects",
  allChats: "All chats",
  create: "Create project",
  createTitle: "Create project",
  editTitle: "Edit project",
  name: "Project name",
  context: "Shared project context",
  contextHint: "Notes Caret can use across chats in this project.",
  icon: "Icon",
  color: "Color",
  save: "Save project",
  cancel: "Cancel",
  archive: "Archive project",
  archiveTitle: "Archive project?",
  archiveMessage:
    "Its chats will remain in All chats and lose this shared project context.",
  confirmArchive: "Archive",
  empty: "Create a project to keep related chats and context together.",
  loadFailed: "Unable to load projects.",
  required: "Enter a project name.",
  loading: "Loading…",
  saving: "Saving…",
  saveFailed: "Unable to save project.",
  archiveFailed: "Unable to archive project.",
};

const ProjectSidebar: React.FC<ProjectSidebarProps> = ({
  projects,
  activeId,
  labels,
  isLoading = false,
  loadError = false,
  onSelect,
  onSave,
  onArchive,
}) => {
  const copy = { ...DEFAULT_LABELS, ...labels };
  const [editing, setEditing] = useState<CaretProject | null | undefined>(
    undefined,
  );
  const [archiving, setArchiving] = useState<CaretProject | null>(null);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const modalRef = useRef<HTMLDivElement>(null);
  const nameRef = useRef<HTMLInputElement>(null);
  useFocusTrap(modalRef, editing !== undefined || archiving !== null);

  useEffect(() => {
    if (editing !== undefined) nameRef.current?.focus();
  }, [editing]);

  const closeDialogs = (): void => {
    if (saving) return;
    setEditing(undefined);
    setArchiving(null);
    setError("");
  };
  useEffect(() => {
    if (editing === undefined && archiving === null) return;
    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.key === "Escape") closeDialogs();
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  });

  const handleSave = async (
    event: React.FormEvent<HTMLFormElement>,
  ): Promise<void> => {
    event.preventDefault();
    if (editing === undefined) return;
    const form = new FormData(event.currentTarget);
    const name = String(form.get("name") ?? "").trim();
    if (!name) {
      setError(copy.required);
      return;
    }
    setSaving(true);
    setError("");
    try {
      const saved = await onSave({
        id: editing?.id ?? "",
        name,
        icon: String(form.get("icon") ?? "folder"),
        color: String(form.get("color") ?? "neutral"),
        context: String(form.get("context") ?? "").trim(),
      });
      setEditing(undefined);
      onSelect(saved.id);
    } catch (saveError) {
      setError(
        saveError instanceof Error ? saveError.message : copy.saveFailed,
      );
    } finally {
      setSaving(false);
    }
  };

  const handleArchive = async (): Promise<void> => {
    if (!archiving) return;
    setSaving(true);
    setError("");
    try {
      await onArchive(archiving.id);
      if (activeId === archiving.id) onSelect(null);
      setArchiving(null);
    } catch (archiveError) {
      setError(
        archiveError instanceof Error
          ? archiveError.message
          : copy.archiveFailed,
      );
    } finally {
      setSaving(false);
    }
  };

  const modalOpen = editing !== undefined || archiving !== null;
  return (
    <>
      <section
        aria-label={copy.heading}
        style={{ padding: "8px 14px 8px", flexShrink: 0 }}
      >
        <h2
          style={{
            margin: 0,
            padding: "4px 2px",
            color: caretTokens.textDim,
            fontSize: 11.5,
            fontWeight: 650,
          }}
        >
          {copy.heading}
        </h2>
        <div style={{ display: "grid", gap: 2 }}>
          {loadError ? (
            <p
              role="alert"
              style={{
                margin: "6px 8px",
                fontSize: 11,
                color: caretTokens.textDim,
              }}
            >
              {copy.loadFailed}
            </p>
          ) : null}
          {isLoading ? (
            <p
              role="status"
              style={{
                margin: "6px 8px",
                fontSize: 11,
                color: caretTokens.textDim,
              }}
            >
              {copy.loading}
            </p>
          ) : null}
          {projects.map((project) => (
            <ProjectRow
              key={project.id}
              name={project.name}
              icon={project.icon}
              color={project.color}
              active={activeId === project.id}
              onSelect={() =>
                onSelect(activeId === project.id ? null : project.id)
              }
              onEdit={() => setEditing(project)}
              onArchive={() => setArchiving(project)}
              editLabel={copy.editTitle}
              archiveLabel={copy.archive}
            />
          ))}
          <button
            type="button"
            onClick={() => setEditing(null)}
            style={{
              display: "flex",
              alignItems: "center",
              gap: 9,
              border: 0,
              borderRadius: caretRadius.sm,
              background: "transparent",
              color: caretTokens.textMuted,
              padding: "6px 9px",
              cursor: "pointer",
              textAlign: "left",
              fontSize: 12.5,
            }}
          >
            <Plus size={15} style={{ flexShrink: 0 }} />
            {copy.create}
          </button>
        </div>
      </section>

      {modalOpen && (
        <div
          className="caret-project-backdrop"
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) closeDialogs();
          }}
          style={{
            position: "fixed",
            inset: 0,
            zIndex: 10001,
            display: "grid",
            placeItems: "center",
            padding: 18,
            background: "rgba(0,0,0,.38)",
          }}
        >
          <div
            ref={modalRef}
            role={archiving ? "alertdialog" : "dialog"}
            aria-modal="true"
            aria-labelledby="caret-project-dialog-title"
            style={{
              width: "min(520px, 100%)",
              maxHeight: "min(700px, 92vh)",
              overflowY: "auto",
              background: caretTokens.bg,
              border: `1px solid ${caretTokens.border}`,
              borderRadius: caretRadius.lg,
              padding: 20,
              boxSizing: "border-box",
              color: caretTokens.text,
              boxShadow: caretTokens.shadow,
            }}
          >
            {archiving ? (
              <>
                <h2
                  id="caret-project-dialog-title"
                  style={{ margin: "0 36px 9px 0", fontSize: 18 }}
                >
                  {copy.archiveTitle}
                </h2>
                <p
                  style={{
                    color: caretTokens.textMuted,
                    fontSize: 13,
                    lineHeight: 1.5,
                  }}
                >
                  {copy.archiveMessage}
                </p>
                <p style={{ fontSize: 13, fontWeight: 600 }}>
                  {archiving.name}
                </p>
              </>
            ) : editing !== undefined ? (
              <form
                key={editing?.id ?? "new"}
                id="caret-project-form"
                onSubmit={(event) => {
                  void handleSave(event);
                }}
              >
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    gap: 12,
                    marginBottom: 16,
                  }}
                >
                  <h2
                    id="caret-project-dialog-title"
                    style={{ margin: 0, fontSize: 18 }}
                  >
                    {editing ? copy.editTitle : copy.createTitle}
                  </h2>
                  <button
                    type="button"
                    aria-label={copy.cancel}
                    onClick={closeDialogs}
                    style={{
                      border: 0,
                      background: "transparent",
                      color: caretTokens.textDim,
                      cursor: "pointer",
                    }}
                  >
                    <X size={18} />
                  </button>
                </div>
                <label style={labelStyle}>
                  {copy.name}
                  <input
                    ref={nameRef}
                    name="name"
                    defaultValue={editing?.name ?? ""}
                    maxLength={120}
                    required
                    style={inputStyle}
                  />
                </label>
                <fieldset style={{ border: 0, margin: "16px 0 0", padding: 0 }}>
                  <legend style={labelStyle}>{copy.icon}</legend>
                  <div
                    style={{
                      display: "flex",
                      flexWrap: "wrap",
                      gap: 6,
                      marginTop: 7,
                    }}
                  >
                    {ICONS.map(([key, Icon]) => (
                      <label
                        key={key}
                        title={key}
                        style={{ cursor: "pointer" }}
                      >
                        <input
                          type="radio"
                          name="icon"
                          value={key}
                          defaultChecked={(editing?.icon ?? "folder") === key}
                          className="caret-visually-hidden"
                        />
                        <span className="caret-project-icon-choice">
                          <Icon size={17} />
                        </span>
                      </label>
                    ))}
                  </div>
                </fieldset>
                <fieldset style={{ border: 0, margin: "14px 0 0", padding: 0 }}>
                  <legend style={labelStyle}>{copy.color}</legend>
                  <div style={{ display: "flex", gap: 10, marginTop: 8 }}>
                    {COLORS.map(([key, color]) => (
                      <label
                        key={key}
                        title={key}
                        style={{ cursor: "pointer" }}
                      >
                        <input
                          type="radio"
                          name="color"
                          value={key}
                          defaultChecked={(editing?.color ?? "neutral") === key}
                          className="caret-visually-hidden"
                        />
                        <span
                          className="caret-project-color-choice"
                          style={{ background: color }}
                        />
                      </label>
                    ))}
                  </div>
                </fieldset>
                <label
                  style={{ ...labelStyle, display: "block", marginTop: 16 }}
                >
                  {copy.context}
                  <span
                    style={{
                      display: "block",
                      marginTop: 4,
                      fontSize: 11.5,
                      lineHeight: 1.4,
                      color: caretTokens.textDim,
                      fontWeight: 400,
                    }}
                  >
                    {copy.contextHint}
                  </span>
                  <textarea
                    name="context"
                    defaultValue={editing?.context ?? ""}
                    maxLength={12000}
                    rows={4}
                    style={{ ...inputStyle, resize: "vertical", marginTop: 8 }}
                  />
                </label>
              </form>
            ) : null}
            {error && (
              <p
                role="alert"
                style={{ color: caretTokens.textMuted, fontSize: 12 }}
              >
                {error}
              </p>
            )}
            <div
              style={{
                display: "flex",
                justifyContent: "flex-end",
                gap: 8,
                marginTop: 18,
              }}
            >
              <button
                type="button"
                disabled={saving}
                onClick={closeDialogs}
                style={secondaryButtonStyle}
              >
                {copy.cancel}
              </button>
              {archiving ? (
                <button
                  type="button"
                  disabled={saving}
                  onClick={() => {
                    void handleArchive();
                  }}
                  style={{ ...primaryButtonStyle, background: "#b42318" }}
                >
                  {saving ? copy.saving : copy.confirmArchive}
                </button>
              ) : (
                <button
                  type="submit"
                  disabled={saving}
                  form="caret-project-form"
                  style={primaryButtonStyle}
                >
                  {saving ? copy.saving : copy.save}
                </button>
              )}
            </div>
          </div>
        </div>
      )}
      <style>{`
        .caret-project-icon-choice,.caret-project-color-choice { display:grid; place-items:center; width:34px; height:34px; border:1px solid ${caretTokens.border}; border-radius:8px; color:${caretTokens.text}; }
        input[type=radio]:checked + .caret-project-icon-choice,input[type=radio]:checked + .caret-project-color-choice { outline:2px solid ${caretTokens.accent}; outline-offset:2px; }
        .caret-project-color-choice { width:20px; height:20px; border-radius:50%; border:0; }
      `}</style>
    </>
  );
};

function ProjectRow({
  name,
  icon,
  color,
  active,
  onSelect,
  onEdit,
  onArchive,
  editLabel,
  archiveLabel,
}: {
  name: string;
  icon: string;
  color: string;
  active: boolean;
  onSelect: () => void;
  onEdit?: () => void;
  onArchive?: () => void;
  editLabel?: string;
  archiveLabel?: string;
}) {
  const Icon = ICONS.find(([key]) => key === icon)?.[1] ?? Folder;
  const iconColor = COLORS.find(([key]) => key === color)?.[1] ?? COLORS[0][1];
  return (
    <div
      className="caret-project-row"
      style={{
        display: "flex",
        alignItems: "center",
        minWidth: 0,
        minHeight: 34,
        borderRadius: caretRadius.sm,
        background: active ? caretTokens.surfaceHigh : "transparent",
      }}
    >
      <button
        type="button"
        onClick={onSelect}
        aria-current={active ? "page" : undefined}
        style={{
          display: "flex",
          alignItems: "center",
          gap: 8,
          flex: 1,
          minWidth: 0,
          border: 0,
          borderRadius: caretRadius.sm,
          background: "transparent",
          color: active ? caretTokens.text : caretTokens.textMuted,
          padding: "6px 9px",
          cursor: "pointer",
          textAlign: "left",
          fontSize: 12.5,
        }}
      >
        <Icon size={15} style={{ flexShrink: 0, color: iconColor }} />
        <span
          style={{
            overflow: "hidden",
            whiteSpace: "nowrap",
            textOverflow: "ellipsis",
          }}
        >
          {name}
        </span>
      </button>
      {onEdit && (
        <button
          type="button"
          aria-label={editLabel}
          title={editLabel}
          onClick={onEdit}
          style={rowActionStyle}
        >
          <Pencil size={12} />
        </button>
      )}
      {onArchive && (
        <button
          type="button"
          aria-label={archiveLabel}
          title={archiveLabel}
          onClick={onArchive}
          style={rowActionStyle}
        >
          <Archive size={13} />
        </button>
      )}
    </div>
  );
}

const labelStyle: React.CSSProperties = {
  display: "block",
  color: caretTokens.text,
  fontSize: 12.5,
  fontWeight: 600,
};
const inputStyle: React.CSSProperties = {
  width: "100%",
  boxSizing: "border-box",
  border: `1px solid ${caretTokens.border}`,
  borderRadius: caretRadius.sm,
  padding: "9px 10px",
  background: caretTokens.surface,
  color: caretTokens.text,
  font: "inherit",
  fontSize: 13,
  marginTop: 7,
};
const secondaryButtonStyle: React.CSSProperties = {
  border: `1px solid ${caretTokens.border}`,
  borderRadius: caretRadius.sm,
  padding: "8px 12px",
  background: "transparent",
  color: caretTokens.text,
  cursor: "pointer",
  fontSize: 12.5,
};
const primaryButtonStyle: React.CSSProperties = {
  border: 0,
  borderRadius: caretRadius.sm,
  padding: "8px 12px",
  background: caretTokens.accent,
  color: caretTokens.accentInk,
  cursor: "pointer",
  fontSize: 12.5,
  fontWeight: 600,
};
const rowActionStyle: React.CSSProperties = {
  display: "grid",
  placeItems: "center",
  width: 24,
  height: 24,
  marginRight: 3,
  border: 0,
  borderRadius: caretRadius.sm,
  background: "transparent",
  color: caretTokens.textDim,
  cursor: "pointer",
};

export default ProjectSidebar;
