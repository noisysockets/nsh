// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 The Noisy Sockets Authors.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

import React, {
  useEffect,
  useMemo,
  useRef,
  useState,
  CSSProperties,
} from "react";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { EventEmitter } from "eventemitter3";

import "@xterm/xterm/css/xterm.css";
import "./XTerm.css";

interface TerminalDimensions {
  columns: number;
  rows: number;
}

interface XTermProps {
  // input is where the terminal will read data from.
  input: EventEmitter;
  // output is where the terminal will write data to.
  output: EventEmitter;
  // style sets the style of the terminal container.
  style?: CSSProperties;
}

// XTerm is a React component that wraps the xterm.js terminal emulator.
const XTerm: React.FC<XTermProps> = ({ input, output, style }) => {
  const textDecoder = useMemo<TextDecoder>(() => new TextDecoder(), []);
  const textEncoder = useMemo<TextEncoder>(() => new TextEncoder(), []);

  const terminalRef = useRef<HTMLDivElement | null>(null);
  const [term, setTerm] = useState<Terminal | null>(null);
  const fitAddon = useMemo(() => new FitAddon(), []);

  useEffect(() => {
    const terminal = new Terminal({
      cursorBlink: true,
    });

    setTerm(terminal);

    return () => {
      terminal.dispose();
    };
  }, []);

  useEffect(() => {
    if (term) {
      const handleResize = () => {
        fitAddon.fit();
        output.emit("resize", {
          columns: term.cols,
          rows: term.rows,
        } as TerminalDimensions);
      };
      window.addEventListener("resize", handleResize);

      if (terminalRef.current) {
        term.loadAddon(fitAddon);
        term.open(terminalRef.current);
        handleResize();
        term.focus();
        setTerm(term);
      }

      const handleInputData = (data: Uint8Array) => {
        term.write(textDecoder.decode(data));
      };
      input.on("data", handleInputData);

      const handleOutputData = term.onData((data) => {
        output.emit("data", textEncoder.encode(data));
      });

      const handleTitleChange = term.onTitleChange((title: string) => {
        output.emit("title", title);
      });

      return () => {
        window.removeEventListener("resize", handleResize);
        input.off("data", handleInputData);
        handleOutputData.dispose();
        handleTitleChange.dispose();
      };
    }
  }, [term, input, output, textDecoder, textEncoder, fitAddon]);

  return <div ref={terminalRef} style={style} />;
};

export default XTerm;
export type { TerminalDimensions };
