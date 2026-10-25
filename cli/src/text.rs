//! Building blocks for the human readable output.

use std::fmt::Write;

use chrono::{DateTime, Utc};

use crate::api::User;

/// A table with aligned columns.
pub struct Table {
    headers: Vec<&'static str>,
    rows: Vec<Vec<String>>,
}

impl Table {
    pub fn new(headers: &[&'static str]) -> Self {
        Self {
            headers: headers.to_vec(),
            rows: Vec::new(),
        }
    }

    pub fn row(&mut self, cells: Vec<String>) {
        self.rows.push(cells);
    }

    /// Renders the table, or `empty` if it has no rows.
    pub fn render(&self, empty: &str) -> String {
        if self.rows.is_empty() {
            return empty.to_owned();
        }
        let mut widths: Vec<usize> = self.headers.iter().map(|header| header.len()).collect();
        for row in &self.rows {
            for (width, cell) in widths.iter_mut().zip(row) {
                *width = (*width).max(cell.chars().count());
            }
        }
        let headers: Vec<String> = self
            .headers
            .iter()
            .map(|header| header.to_string())
            .collect();
        let mut out = String::new();
        for row in std::iter::once(&headers).chain(&self.rows) {
            let mut line = String::new();
            for (i, (cell, width)) in row.iter().zip(&widths).enumerate() {
                if i + 1 < row.len() {
                    let _ = write!(line, "{cell:<width$}  ");
                } else {
                    line.push_str(cell);
                }
            }
            if !out.is_empty() {
                out.push('\n');
            }
            out.push_str(line.trim_end());
        }
        out
    }
}

/// Renders `label: value` lines with aligned values. Multi-line values are
/// indented to the value column; empty values are left out.
pub fn fields(fields: &[(&str, String)]) -> String {
    let width = fields
        .iter()
        .map(|(label, _)| label.len() + 1)
        .max()
        .unwrap_or(0);
    let mut out = String::new();
    for (label, value) in fields {
        if value.is_empty() {
            continue;
        }
        let mut lines = value.lines();
        let first = lines.next().unwrap_or_default();
        let _ = writeln!(out, "{:<width$}  {first}", format!("{label}:"));
        for line in lines {
            let _ = writeln!(out, "{:<width$}  {line}", "");
        }
    }
    out
}

/// A footer for paginated results. Pages count from 0, like the API.
pub fn page_footer(page: u64, total_pages: u64, total: u64, (one, many): (&str, &str)) -> String {
    let noun = if total == 1 { one } else { many };
    if total_pages <= 1 {
        return format!("{total} {noun}");
    }
    format!("{total} {noun}, page {page} of 0-{}", total_pages - 1)
}

pub fn date(date: &DateTime<Utc>) -> String {
    date.format("%Y-%m-%d %H:%M UTC").to_string()
}

pub fn user(user: &User) -> String {
    format!("{} ({})", user.name, user.id)
}
