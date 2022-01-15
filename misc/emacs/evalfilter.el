;;; evalfilter.el --- mode for editing evalfilter scripts

;; Copyright (C) 2022 Steve Kemp

;; Author: Steve Kemp <steve@steve.fi>
;; Keywords: languages
;; Version: 1.0

;;; Commentary:

;; Provides support for editing scripts with full support for
;; font-locking, but no special keybindings, or indentation handling.

;;;; Enabling:

;; Add the following to your .emacs file

;; (require 'evalfilter)
;; (setq auto-mode-alist (append '(("\\.evalfilter$" . evalfilter-mode)) auto-mode-alist)))



;;; Code:

(defvar evalfilter-constants
  '("true"
    "false"
    "nil"))

(defvar evalfilter-keywords
  '(
    "case"
    "default"
    "else"
    "for"
    "foreach"
    "function"
    "if"
    "in"
    "local"
    "return"
    "switch"
    "while"
    ))

;; The language-core and functions from the standard-library.
(defvar evalfilter-functions
  '(
    "between"
    "day"
    "float"
    "getenv"
    "hour"
    "int"
    "keys"
    "len"
    "lower"
    "match"
    "max"
    "min"
    "minute"
    "month"
    "now"
    "panic"
    "printf"
    "print"
    "reverse"
    "seconds"
    "sort"
    "split"
    "sprintf"
    "string"
    "time"
    "trim"
    "type"
    "upper"
    "weekday"
    "year"
    ))


(defvar evalfilter-font-lock-defaults
  `((
     ("\"\\.\\*\\?" . font-lock-string-face)
     (";\\|,\\|=" . font-lock-keyword-face)
     ( ,(regexp-opt evalfilter-keywords 'words) . font-lock-builtin-face)
     ( ,(regexp-opt evalfilter-constants 'words) . font-lock-constant-face)
     ( ,(regexp-opt evalfilter-functions 'words) . font-lock-function-name-face)
     )))

(define-derived-mode evalfilter-mode fundamental-mode "evalfilter script"
  "evalfilter-mode is a major mode for editing evalfilter scripts"
  (setq font-lock-defaults evalfilter-font-lock-defaults)

  ;; Comment handler for single & multi-line modes
  (modify-syntax-entry ?\/ ". 124b" evalfilter-mode-syntax-table)
  (modify-syntax-entry ?\* ". 23n" evalfilter-mode-syntax-table)

  ;; Comment ender for single-line comments.
  (modify-syntax-entry ?\n "> b" evalfilter-mode-syntax-table)
  (modify-syntax-entry ?\r "> b" evalfilter-mode-syntax-table)
  )

(provide 'evalfilter)
