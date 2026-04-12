# LaTeX Templates for Resume

This directory contains LaTeX templates for resumes.

## Default Resume Template

```latex
\documentclass[11pt]{article}
\usepackage[utf8]{inputenc}
\usepackage[T1]{fontenc}
\usepackage{geometry}
\geometry{a4paper, margin=1in}
\usepackage{hyperref}
\usepackage{graphicx}
\usepackage{parskip}

\title{Your Resume Title}
\author{Your Name}
\date{\today}

\begin{document}

\maketitle

\section{Education}
\begin{itemize}
    \item Degree in Field, University Name, Year
\end{itemize}

\section{Experience}
\begin{itemize}
    \item \textbf{Job Title} - Company Name
    \begin{itemize}
        \item Description of responsibilities and achievements
    \end{itemize}
\end{itemize}

\section{Skills}
\begin{itemize}
    \item Programming Languages: Python, JavaScript, Go
    \item Tools: Git, Docker, Linux
\end{itemize}

\end{document}
```
