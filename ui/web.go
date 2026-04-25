package ui

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sort"
	"sync"
	"terraviz/model"
)

type webGraphServer struct {
	once     sync.Once
	mu       sync.RWMutex
	baseURL  string
	startErr error
	snapshot webGraphSnapshot
}

type webGraphSnapshot struct {
	Provider string         `json:"provider"`
	Nodes    []webGraphNode `json:"nodes"`
	Edges    []webGraphEdge `json:"edges"`
}

type webGraphNode struct {
	ID            string              `json:"id"`
	Label         string              `json:"label"`
	Type          string              `json:"type"`
	Provider      string              `json:"provider"`
	Module        string              `json:"module"`
	Dependencies  []string            `json:"dependencies"`
	LocalRefs     []string            `json:"localRefs"`
	VariableRefs  []string            `json:"variableRefs"`
	ModuleInputs  map[string][]string `json:"moduleInputs"`
	AttributeKeys []string            `json:"attributeKeys"`
}

type webGraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func newWebGraphServer() *webGraphServer {
	return &webGraphServer{}
}

func (s *webGraphServer) Open(resources map[string]*model.Resource, edges []model.Edge, provider string) (string, error) {
	s.mu.Lock()
	s.snapshot = buildWebGraphSnapshot(resources, edges, provider)
	s.mu.Unlock()

	if err := s.start(); err != nil {
		return "", err
	}

	url := s.baseURL
	if err := openBrowser(url); err != nil {
		return "", err
	}

	return url, nil
}

func (s *webGraphServer) start() error {
	s.once.Do(func() {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			s.startErr = err
			return
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/", s.handleIndex)
		mux.HandleFunc("/graph.json", s.handleGraph)

		s.baseURL = "http://" + listener.Addr().String()

		go func() {
			server := &http.Server{Handler: mux}
			_ = server.Serve(listener)
		}()
	})

	return s.startErr
}

func (s *webGraphServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(webGraphHTML))
}

func (s *webGraphServer) handleGraph(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	snapshot := s.snapshot
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(snapshot)
}

func buildWebGraphSnapshot(resources map[string]*model.Resource, edges []model.Edge, provider string) webGraphSnapshot {
	nodes := make([]webGraphNode, 0, len(resources))
	for _, resource := range resources {
		attributeKeys := make([]string, 0, len(resource.Attributes))
		for key := range resource.Attributes {
			attributeKeys = append(attributeKeys, key)
		}
		sort.Strings(attributeKeys)

		nodes = append(nodes, webGraphNode{
			ID:            resource.ID,
			Label:         fmt.Sprintf("%s.%s", resource.Type, resource.Name),
			Type:          resource.Type,
			Provider:      resource.Provider,
			Module:        resource.Module,
			Dependencies:  append([]string(nil), resource.Dependencies...),
			LocalRefs:     append([]string(nil), resource.LocalRefs...),
			VariableRefs:  append([]string(nil), resource.VariableRefs...),
			ModuleInputs:  cloneModuleInputs(resource.ModuleInputs),
			AttributeKeys: attributeKeys,
		})
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Label < nodes[j].Label
	})

	webEdges := make([]webGraphEdge, 0, len(edges))
	for _, edge := range edges {
		webEdges = append(webEdges, webGraphEdge{From: edge.From, To: edge.To})
	}

	return webGraphSnapshot{
		Provider: provider,
		Nodes:    nodes,
		Edges:    webEdges,
	}
}

func cloneModuleInputs(inputs map[string][]string) map[string][]string {
	if len(inputs) == 0 {
		return map[string][]string{}
	}

	cloned := make(map[string][]string, len(inputs))
	for key, values := range inputs {
		cloned[key] = append([]string(nil), values...)
	}

	return cloned
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

const webGraphHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Terraviz Graph</title>
  <style>
    :root {
      --bg: #0b1220;
      --panel: #121a2b;
      --panel-2: #172033;
      --text: #eef2ff;
      --muted: #98a4c1;
      --line: #33415f;
      --aws: #f59e0b;
      --google: #60a5fa;
      --azure: #22d3ee;
      --accent: #8b5cf6;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background:
        radial-gradient(circle at top left, rgba(139, 92, 246, 0.18), transparent 28%),
        radial-gradient(circle at bottom right, rgba(34, 211, 238, 0.14), transparent 30%),
        var(--bg);
      color: var(--text);
    }
    .app {
      height: 100vh;
      display: grid;
      grid-template-columns: minmax(0, 1fr) 320px;
      gap: 16px;
      padding: 16px;
    }
    .canvas-panel, .side-panel {
      background: rgba(18, 26, 43, 0.92);
      border: 1px solid rgba(139, 92, 246, 0.35);
      border-radius: 18px;
      overflow: hidden;
      box-shadow: 0 20px 50px rgba(0, 0, 0, 0.35);
      min-height: 0;
    }
    .header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 12px;
      padding: 14px 18px;
      border-bottom: 1px solid rgba(255,255,255,0.08);
      background: rgba(255,255,255,0.02);
    }
    .title {
      font-size: 18px;
      font-weight: 700;
      letter-spacing: 0.02em;
    }
    .subtitle {
      font-size: 12px;
      color: var(--muted);
      margin-top: 4px;
    }
    .badge {
      padding: 6px 10px;
      border-radius: 999px;
      background: rgba(139, 92, 246, 0.16);
      border: 1px solid rgba(139, 92, 246, 0.4);
      color: #ddd6fe;
      font-size: 12px;
      white-space: nowrap;
    }
    .canvas-wrap {
      height: calc(100% - 61px);
      overflow: auto;
      background-image:
        linear-gradient(rgba(255,255,255,0.04) 1px, transparent 1px),
        linear-gradient(90deg, rgba(255,255,255,0.04) 1px, transparent 1px);
      background-size: 28px 28px;
    }
    svg {
      display: block;
      min-width: 100%;
      min-height: 100%;
    }
    .edge {
      stroke: rgba(148, 163, 184, 0.45);
      stroke-width: 2;
    }
    .edge-label {
      fill: rgba(148, 163, 184, 0.85);
      font-size: 11px;
    }
    .node rect {
      stroke-width: 1.5;
      rx: 14;
    }
    .node text.label {
      fill: var(--text);
      font-size: 13px;
      font-weight: 600;
    }
    .node text.meta {
      fill: var(--muted);
      font-size: 11px;
    }
    .node.selected rect {
      stroke: #e9d5ff;
      stroke-width: 2.5;
      filter: drop-shadow(0 0 12px rgba(139, 92, 246, 0.45));
    }
    .side-panel {
      display: flex;
      flex-direction: column;
    }
    .side-content {
      padding: 16px 18px 20px;
      overflow: auto;
      display: grid;
      gap: 14px;
    }
    .section-title {
      font-size: 11px;
      letter-spacing: 0.12em;
      text-transform: uppercase;
      color: #c4b5fd;
      margin-bottom: 8px;
    }
    .kv {
      display: grid;
      grid-template-columns: 92px minmax(0, 1fr);
      gap: 8px;
      margin-bottom: 8px;
      font-size: 13px;
    }
    .kv .key { color: var(--muted); }
    ul.clean {
      margin: 0;
      padding-left: 18px;
      color: var(--text);
      display: grid;
      gap: 6px;
      font-size: 13px;
    }
    .legend {
      display: flex;
      gap: 10px;
      flex-wrap: wrap;
      font-size: 12px;
      color: var(--muted);
    }
    .legend span::before {
      content: "";
      display: inline-block;
      width: 10px;
      height: 10px;
      border-radius: 999px;
      margin-right: 6px;
      vertical-align: middle;
    }
    .legend .aws::before { background: var(--aws); }
    .legend .google::before { background: var(--google); }
    .legend .azurerm::before { background: var(--azure); }
    .legend .other::before { background: var(--accent); }
    .empty {
      color: var(--muted);
      font-size: 13px;
      line-height: 1.5;
    }
    @media (max-width: 980px) {
      .app { grid-template-columns: 1fr; }
      .side-panel { min-height: 260px; }
    }
  </style>
</head>
<body>
  <div class="app">
    <section class="canvas-panel">
      <div class="header">
        <div>
          <div class="title">Terraviz Web Graph</div>
          <div id="graph-subtitle" class="subtitle">Loading graph…</div>
        </div>
        <div id="graph-provider" class="badge">provider</div>
      </div>
      <div class="canvas-wrap">
        <svg id="graph" viewBox="0 0 1600 1000" preserveAspectRatio="xMinYMin meet"></svg>
      </div>
    </section>
    <aside class="side-panel">
      <div class="header">
        <div>
          <div class="title">Selection</div>
          <div class="subtitle">Click any node in the diagram</div>
        </div>
      </div>
      <div id="details" class="side-content">
        <div>
          <div class="section-title">Legend</div>
          <div class="legend">
            <span class="aws">AWS</span>
            <span class="google">Google</span>
            <span class="azurerm">Azure</span>
            <span class="other">Other</span>
          </div>
        </div>
        <div class="empty">Select a resource to inspect its provider, module, locals, variables, and dependencies.</div>
      </div>
    </aside>
  </div>
  <script>
    const providerColors = {
      aws: '#f59e0b',
      google: '#60a5fa',
      gcp: '#60a5fa',
      azurerm: '#22d3ee',
      azure: '#22d3ee'
    };

    fetch('/graph.json')
      .then((response) => response.json())
      .then((graph) => renderGraph(graph))
      .catch((error) => {
        document.getElementById('graph-subtitle').textContent = 'Failed to load graph';
        document.getElementById('details').innerHTML = '<div class="empty">' + error + '</div>';
      });

    function renderGraph(graph) {
      const svg = document.getElementById('graph');
      const subtitle = document.getElementById('graph-subtitle');
      const providerBadge = document.getElementById('graph-provider');
      const nodeMap = new Map(graph.nodes.map((node) => [node.id, node]));
      const children = new Map();
      const parentMap = new Map();

      providerBadge.textContent = graph.provider || 'all providers';

      for (const node of graph.nodes) {
        children.set(node.id, []);
      }

      for (const edge of graph.edges) {
        if (!children.has(edge.to)) children.set(edge.to, []);
        children.get(edge.to).push(edge.from);
        if (!parentMap.has(edge.from)) {
          parentMap.set(edge.from, edge.to);
        }
      }

      for (const childList of children.values()) {
        childList.sort((left, right) => {
          const leftNode = nodeMap.get(left);
          const rightNode = nodeMap.get(right);
          return (leftNode ? leftNode.label : left).localeCompare(rightNode ? rightNode.label : right);
        });
      }

      const rootNodes = graph.nodes
        .filter((node) => (node.dependencies || []).length === 0)
        .sort((left, right) => left.label.localeCompare(right.label));
      const fallbackRoot = rootNodes[0] || graph.nodes[0] || null;
      let selectedNodeId = fallbackRoot ? fallbackRoot.id : null;

      function getBranch(nodeId) {
        const branch = [];
        const seen = new Set();
        let current = nodeId;
        while (current && !seen.has(current) && nodeMap.has(current)) {
          branch.unshift(current);
          seen.add(current);
          current = parentMap.get(current) || null;
        }
        return branch;
      }

      function getVisibleColumns(nodeId) {
        const branch = getBranch(nodeId);
        const columns = [];
        columns.push(rootNodes.length > 0 ? rootNodes : (fallbackRoot ? [fallbackRoot] : []));
        for (const branchNodeId of branch) {
          const childIds = children.get(branchNodeId) || [];
          if (childIds.length === 0) {
            continue;
          }
          columns.push(childIds.map((childId) => nodeMap.get(childId)).filter(Boolean));
        }
        return columns.filter((column) => column.length > 0);
      }

      function draw() {
        const columns = getVisibleColumns(selectedNodeId);
        const visibleCount = columns.reduce((count, column) => count + column.length, 0);
        subtitle.textContent = String(visibleCount) + ' visible of ' + String(graph.nodes.length) + ' resources • expand by selecting parents';

        const levelGap = 310;
        const rowGap = 132;
        const nodeWidth = 240;
        const nodeHeight = 76;
        const margin = 80;
        const width = margin * 2 + Math.max(columns.length, 1) * levelGap;
        const tallestColumn = Math.max(1, ...columns.map((column) => column.length));
        const height = margin * 2 + tallestColumn * rowGap;
        const positions = new Map();

        svg.setAttribute('viewBox', '0 0 ' + String(width) + ' ' + String(height));
        svg.innerHTML = '';

        const edgeLayer = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        const nodeLayer = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        svg.appendChild(edgeLayer);
        svg.appendChild(nodeLayer);

        columns.forEach((column, level) => {
          column.forEach((node, index) => {
            positions.set(node.id, {
              x: margin + level * levelGap,
              y: margin + index * rowGap,
            });
          });
        });

        for (let level = 0; level < columns.length - 1; level += 1) {
          for (const node of columns[level]) {
            const from = positions.get(node.id);
            if (!from) {
              continue;
            }
            const childIds = children.get(node.id) || [];
            for (const childId of childIds) {
              const to = positions.get(childId);
              if (!to) {
                continue;
              }
              const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
              const startX = from.x + nodeWidth;
              const startY = from.y + nodeHeight / 2;
              const endX = to.x;
              const endY = to.y + nodeHeight / 2;
              const midX = (startX + endX) / 2;
              path.setAttribute('d', 'M ' + startX + ' ' + startY + ' C ' + midX + ' ' + startY + ', ' + midX + ' ' + endY + ', ' + endX + ' ' + endY);
              path.setAttribute('class', 'edge');
              path.setAttribute('fill', 'none');
              edgeLayer.appendChild(path);
            }
          }
        }

        for (const column of columns) {
          for (const node of column) {
            const pos = positions.get(node.id);
            if (!pos) {
              continue;
            }
            const group = document.createElementNS('http://www.w3.org/2000/svg', 'g');
            group.setAttribute('class', 'node' + (node.id === selectedNodeId ? ' selected' : ''));
            group.dataset.nodeId = node.id;
            group.setAttribute('transform', 'translate(' + pos.x + ', ' + pos.y + ')');

            const rect = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
            rect.setAttribute('width', String(nodeWidth));
            rect.setAttribute('height', String(nodeHeight));
            rect.setAttribute('fill', 'rgba(23, 32, 51, 0.95)');
            rect.setAttribute('stroke', providerColors[node.provider] || '#8b5cf6');
            group.appendChild(rect);

            const label = document.createElementNS('http://www.w3.org/2000/svg', 'text');
            label.setAttribute('x', '14');
            label.setAttribute('y', '28');
            label.setAttribute('class', 'label');
            label.textContent = truncate(node.label, 28);
            group.appendChild(label);

            const childCount = (children.get(node.id) || []).length;
            const meta = document.createElementNS('http://www.w3.org/2000/svg', 'text');
            meta.setAttribute('x', '14');
            meta.setAttribute('y', '50');
            meta.setAttribute('class', 'meta');
            meta.textContent = (node.provider || 'unknown') + ' • ' + String(childCount) + ' children';
            group.appendChild(meta);

            group.addEventListener('click', () => {
              selectedNodeId = node.id;
              renderDetails(node);
              draw();
            });

            nodeLayer.appendChild(group);
          }
        }
      }

      if (fallbackRoot) {
        renderDetails(fallbackRoot);
        draw();
      } else {
        subtitle.textContent = 'No resources available';
      }
    }

    function renderDetails(node) {
      const details = document.getElementById('details');
      const sections = [];

      sections.push(
        '<div>' +
          '<div class="section-title">Resource</div>' +
          '<div class="kv"><div class="key">ID</div><div>' + escapeHtml(node.id) + '</div></div>' +
          '<div class="kv"><div class="key">Type</div><div>' + escapeHtml(node.type) + '</div></div>' +
          '<div class="kv"><div class="key">Provider</div><div>' + escapeHtml(node.provider || 'unknown') + '</div></div>' +
          '<div class="kv"><div class="key">Module</div><div>' + escapeHtml(node.module || 'root') + '</div></div>' +
        '</div>'
      );

      sections.push(renderListSection('Dependencies', node.dependencies || []));
      sections.push(renderListSection('Variables', node.variableRefs || []));
      sections.push(renderListSection('Locals', node.localRefs || []));
      sections.push(renderMapSection('Module Inputs', node.moduleInputs || {}));
      sections.push(renderListSection('Attributes', node.attributeKeys || []));

      details.innerHTML = sections.filter(Boolean).join('');
    }

    function renderListSection(title, items) {
      if (!items || items.length === 0) return '';
      return '<div>' +
        '<div class="section-title">' + escapeHtml(title) + '</div>' +
        '<ul class="clean">' + items.map((item) => '<li>' + escapeHtml(item) + '</li>').join('') + '</ul>' +
      '</div>';
    }

    function renderMapSection(title, entries) {
      const keys = Object.keys(entries || {});
      if (keys.length === 0) return '';
      keys.sort();
      return '<div>' +
        '<div class="section-title">' + escapeHtml(title) + '</div>' +
        keys.map((key) =>
          '<div class="kv">' +
            '<div class="key">' + escapeHtml(key) + '</div>' +
            '<div>' + escapeHtml((entries[key] || []).join(', ')) + '</div>' +
          '</div>'
        ).join('') +
      '</div>';
    }

    function truncate(value, max) {
      return value.length > max ? value.slice(0, max - 3) + '...' : value;
    }

    function escapeHtml(value) {
      return String(value)
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;');
    }
  </script>
</body>
</html>`
