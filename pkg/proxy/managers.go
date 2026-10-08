package proxy

import "github.com/corlin/AIMeter/pkg/registry"

// SetRegistry binds the unified control plane registry to ProxyHandler.
// Individual engines are reachable as promoted fields (e.g. h.BudgetManager).
func (h *ProxyHandler) SetRegistry(reg *registry.ControlPlaneRegistry) {
	if reg != nil {
		h.ControlPlaneRegistry = reg
	}
}
