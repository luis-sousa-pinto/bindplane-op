import React, { ChangeEvent, useState } from "react";
import "@uiw/react-textarea-code-editor/dist.css";
import CodeEditor from "@uiw/react-textarea-code-editor";
import { classes } from "../../utils/styles";
import { useRef } from "react";
import { useEffect } from "react";
import { ExpandButton } from "../ExpandButton";

import styles from "./yaml-editor.module.scss";

interface YamlEditorProps
  extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  readOnly?: boolean;
  limitHeight?: boolean;
  minHeight?: number;
  onValueChange?: (e: ChangeEvent<HTMLTextAreaElement>) => void;
  value: string;
  actions?: JSX.Element | null;
}
/**
 * YamlEditor is a component that renders a textarea with syntax highlighting.
 * The value is controlled and can be altered with onValueChange prop.
 * If actions prop is defined it will display the actions in the top right corner.
 */
export const YamlEditor: React.FC<YamlEditorProps> = ({
  limitHeight = false,
  readOnly,
  value,
  onValueChange,
  actions,
  minHeight,
  ...rest
}) => {
  // We are only using light theme right now.  This overrides the styling if a user
  // has dark mode as a browser preference.
  document.documentElement.setAttribute("data-color-mode", "light");

  const ref = useRef<HTMLTextAreaElement | null>(null);

  const [expanded, setExpanded] = useState(false);
  const [expandable, setExpandable] = useState(false);

  useEffect(() => {
    if (ref.current && ref.current.scrollHeight > 500) {
      setExpandable(true);
      return;
    }
    setExpandable(false);
  }, [ref]);

  const codeEditorClasses = [
    styles.editor,
    expanded || !limitHeight || !readOnly ? styles.expanded : undefined,
    !readOnly ? styles["bg-white"] : undefined,
  ];

  if (expanded || !limitHeight || !readOnly) {
    codeEditorClasses.push(styles.expanded);
  }

  return (
    <div
      className={rest.disabled || readOnly ? styles.container : styles.editable}
    >
      {actions && <div className={styles.actions}>{actions}</div>}
      <CodeEditor
        {...rest}
        data-testid="yaml-editor"
        className={classes(codeEditorClasses)}
        readOnly={readOnly}
        value={value}
        ref={ref}
        language="yaml"
        onChange={onValueChange}
        padding={15}
        minHeight={minHeight}
      />
      {/* Allow expand and collapse in readOnly mode, if height is over 300 and the limitHeight prop was passed */}
      {readOnly && expandable && limitHeight && (
        <div className={styles["btn-container"]}>
          <ExpandButton
            expanded={expanded}
            onToggleExpanded={() => setExpanded((prev) => !prev)}
          />
        </div>
      )}
    </div>
  );
};
