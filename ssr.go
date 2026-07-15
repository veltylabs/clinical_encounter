//go:build !wasm

package clinical_encounter

// RenderCSS returns embedded CSS for all components.
func RenderCSS() string {
	return `
.ce-patient-head {
  font-weight: bold;
  font-size: 1.1rem;
  color: #333;
}

.ce-form-container {
  padding: 1rem;
  background-color: #f9f9f9;
}

.ce-history-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.ce-history-card {
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  background-color: #f9f9f9;
  cursor: pointer;
  transition: all 0.2s ease;
}

.ce-history-card:hover {
  background-color: #f0f0f0;
  border-color: #007bff;
}

.ce-card-date {
  font-weight: 600;
  font-size: 0.9rem;
  color: #333;
}

.ce-card-status {
  font-size: 0.8rem;
  padding: 0.25rem 0.5rem;
  border-radius: 3px;
  display: inline-block;
  margin-top: 0.25rem;
  font-weight: 600;
}

.ce-status-created {
  background-color: #e9ecef;
  color: #495057;
}

.ce-status-arrived {
  background-color: #d1ecf1;
  color: #0c5460;
}

.ce-status-triaged {
  background-color: #fff3cd;
  color: #856404;
}

.ce-status-in_progress {
  background-color: #ffe0b2;
  color: #bf6a00;
}

.ce-status-completed {
  background-color: #d4edda;
  color: #155724;
}

.ce-status-cancelled {
  background-color: #f8d7da;
  color: #721c24;
}

.ce-card-reason {
  font-size: 0.85rem;
  color: #666;
  margin-top: 0.25rem;
}

.ce-search-container {
  padding: 0.5rem;
}
`
}
