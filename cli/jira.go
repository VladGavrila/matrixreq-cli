package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/config"
	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/VladGavrila/matrixreq-cli/internal/service"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(jiraCmd)
	jiraCmd.AddCommand(jiraInitCmd)
	jiraCmd.AddCommand(jiraGetCmd)
	jiraCmd.AddCommand(jiraAddCmd)
	jiraCmd.AddCommand(jiraRemoveCmd)

	jiraAddCmd.Flags().StringSlice("issue", nil, "Jira issue key (repeatable, e.g. ABC-123)")
	jiraAddCmd.Flags().StringSlice("title", nil, "Jira issue title (repeatable, 1:1 with --issue; defaults to the issue key)")
	jiraAddCmd.Flags().String("jira-base", "", "Jira base URL (overrides the value saved via 'mxreq jira init')")
	jiraAddCmd.Flags().Int("plugin-id", api.JiraPluginIDDefault, "Matrix plugin ID for the Jira add-on")
	jiraAddCmd.Flags().BoolP("yes", "y", false, "Skip the confirmation prompt")
	jiraAddCmd.Flags().Bool("skip-validate", false, "Skip the ghost-item sanity check")
	jiraAddCmd.Flags().StringP("reason", "r", "", "Reason for the change (required, kept in local audit output)")
	_ = jiraAddCmd.MarkFlagRequired("issue")
	_ = jiraAddCmd.MarkFlagRequired("reason")

	jiraRemoveCmd.Flags().StringSlice("issue", nil, "Jira issue key to unlink (repeatable). Optional — if omitted, items that have exactly one Jira link are auto-unlinked")
	jiraRemoveCmd.Flags().Int("plugin-id", api.JiraPluginIDDefault, "Matrix plugin ID for the Jira add-on")
	jiraRemoveCmd.Flags().BoolP("yes", "y", false, "Skip the confirmation prompt")
	jiraRemoveCmd.Flags().StringP("reason", "r", "", "Reason for the change (required, kept in local audit output)")
	_ = jiraRemoveCmd.MarkFlagRequired("reason")
}

var jiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Manage Jira issue links on Matrix items",
	Long: `Manage external Jira-issue links on Matrix items via the /rest/2/wfgw/ API.

This is for external Matrix <-> Jira traceability links only. For internal
Matrix uplinks/downlinks between items, use ` + "`mxreq item link`" + ` instead.

Run 'mxreq jira init' once to configure the Jira base URL for your instance.
Not every Matrix project has the Jira add-on enabled — projects without it
will return HTTP 500 from the wfgw endpoint.`,
}

var jiraInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Configure the Jira base URL used when linking issues",
	Long: `Prompt for the Jira base URL and persist it to the mxreq config file.

The Jira base URL is used to construct browse URLs when linking Jira issues
to Matrix items (e.g. https://example.atlassian.net/browse/ABC-123).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		reader := bufio.NewReader(os.Stdin)

		current := cfg.JiraBaseURL
		if current != "" {
			fmt.Printf("Jira base URL [%s]: ", current)
		} else {
			fmt.Print("Jira base URL (e.g. https://example.atlassian.net): ")
		}
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			line = current
		}
		if line == "" {
			return fmt.Errorf("Jira base URL is required")
		}
		if err := validateJiraBase(line); err != nil {
			return err
		}
		cfg.JiraBaseURL = strings.TrimRight(line, "/")

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		path, _ := config.ConfigPath()
		fmt.Printf("Jira base URL saved to %s\n", path)
		return nil
	},
}

// validateJiraBase rejects obviously malformed Jira base URLs.
func validateJiraBase(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid Jira base URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid Jira base URL: expected http(s) scheme")
	}
	if u.Host == "" {
		return fmt.Errorf("invalid Jira base URL: missing host")
	}
	return nil
}

// resolveJiraBase returns the effective Jira base URL: the flag if set, else
// the value saved in config. Empty means the user hasn't configured one.
func resolveJiraBase(flagVal string) (string, error) {
	if flagVal != "" {
		if err := validateJiraBase(flagVal); err != nil {
			return "", err
		}
		return strings.TrimRight(flagVal, "/"), nil
	}
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}
	if cfg.JiraBaseURL == "" {
		return "", fmt.Errorf("Jira base URL is not set — run 'mxreq jira init' or pass --jira-base")
	}
	return strings.TrimRight(cfg.JiraBaseURL, "/"), nil
}

// requireJiraConfigured is a preflight check for commands that don't consume
// the Jira base URL directly (get, remove) but should still fail fast when
// Jira integration has not been configured for this mxreq install.
func requireJiraConfigured() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.JiraBaseURL == "" {
		return fmt.Errorf("Jira integration is not configured — run 'mxreq jira init' to set the Jira base URL")
	}
	return nil
}

var jiraGetCmd = &cobra.Command{
	Use:   "get <item-ref> [<item-ref>...]",
	Short: "List Jira issues linked to one or more Matrix items",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireJiraConfigured(); err != nil {
			return err
		}
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}

		type row struct {
			Item      string            `json:"item"`
			Project   string            `json:"project"`
			JiraKey   string            `json:"jiraKey"`
			IssueType string            `json:"issueType"`
			Status    string            `json:"status"`
			Title     string            `json:"title"`
			LinkedOn  string            `json:"linkedOn"`
			URL       string            `json:"url"`
			Raw       api.JiraLinkedIssue `json:"-"`
		}

		var all []row
		for _, rawRef := range args {
			itemRef := upperRef(rawRef)
			links, err := svc.Jira.GetLinks(project, itemRef)
			if err != nil {
				return fmt.Errorf("getting Jira links for %s: %w", itemRef, err)
			}
			if len(links) == 0 {
				all = append(all, row{Item: itemRef, Project: project})
				continue
			}
			for _, l := range links {
				meta := decodeJiraMeta(l.ExternalMeta)
				all = append(all, row{
					Item:      itemRef,
					Project:   project,
					JiraKey:   l.ExternalItemID,
					IssueType: meta.IssueType,
					Status:    meta.Status,
					Title:     l.ExternalItemTitle,
					LinkedOn:  shortDate(l.ExternalLinkCreationDate),
					URL:       l.ExternalItemURL,
					Raw:       l,
				})
			}
		}

		if getOutputFormat() == "json" {
			// Drop placeholder rows (items with no links) from the JSON shape —
			// JSON consumers want a straight list of links. Emit one object per
			// item with its Links array.
			type jsonItem struct {
				Item    string                `json:"item"`
				Project string                `json:"project"`
				Links   []api.JiraLinkedIssue `json:"links"`
			}
			grouped := map[string]*jsonItem{}
			var order []string
			for _, r := range all {
				if _, ok := grouped[r.Item]; !ok {
					grouped[r.Item] = &jsonItem{Item: r.Item, Project: r.Project, Links: []api.JiraLinkedIssue{}}
					order = append(order, r.Item)
				}
				if r.JiraKey != "" {
					grouped[r.Item].Links = append(grouped[r.Item].Links, r.Raw)
				}
			}
			out := make([]jsonItem, 0, len(order))
			for _, k := range order {
				out = append(out, *grouped[k])
			}
			return output.PrintItem(getOutputFormat(), out)
		}

		headers := []string{"Item", "Jira Key", "Type", "Status", "Title", "Linked"}
		var rows [][]string
		for _, r := range all {
			if r.JiraKey == "" {
				rows = append(rows, []string{r.Item, "(no links)", "", "", "", ""})
				continue
			}
			rows = append(rows, []string{r.Item, r.JiraKey, r.IssueType, r.Status, truncate(r.Title, 60), r.LinkedOn})
		}
		return output.Print(getOutputFormat(), headers, rows)
	},
}

var jiraAddCmd = &cobra.Command{
	Use:   "add <item-ref> [<item-ref>...]",
	Short: "Link one or more Jira issues to one or more Matrix items",
	Long: `Link Jira issues to Matrix items for traceability.

With multiple --issue flags and multiple item refs, every issue is linked to
every item (cross-product).

Title resolution: if --title is omitted, the Jira issue key is used as the
placeholder title. Pass --title (repeatable, 1:1 with --issue) when you have a
better value from the Jira MCP or UI.

The --reason is required by convention but not sent to the wfgw endpoint.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}

		issues, _ := cmd.Flags().GetStringSlice("issue")
		titles, _ := cmd.Flags().GetStringSlice("title")
		jiraBaseFlag, _ := cmd.Flags().GetString("jira-base")
		pluginID, _ := cmd.Flags().GetInt("plugin-id")
		skipConfirm, _ := cmd.Flags().GetBool("yes")
		skipValidate, _ := cmd.Flags().GetBool("skip-validate")
		reason, _ := cmd.Flags().GetString("reason")

		if len(titles) > 0 && len(titles) != len(issues) {
			return fmt.Errorf("when --title is supplied it must be repeated once per --issue (got %d issues, %d titles)", len(issues), len(titles))
		}
		jiraBase, err := resolveJiraBase(jiraBaseFlag)
		if err != nil {
			return err
		}

		// Build the per-issue template (id/title/url/plugin). matrixItemIds is
		// filled in per-item later.
		templates := make([]api.JiraExternalItem, len(issues))
		for i, key := range issues {
			title := key
			if i < len(titles) && titles[i] != "" {
				title = titles[i]
			}
			templates[i] = api.JiraExternalItem{
				ExternalItemID:    key,
				ExternalItemTitle: title,
				ExternalItemURL:   fmt.Sprintf("%s/browse/%s", jiraBase, key),
				Plugin:            pluginID,
			}
		}

		// Normalize item refs.
		items := make([]string, len(args))
		for i, a := range args {
			items[i] = upperRef(a)
		}

		// Ghost-item validation (inline, skippable).
		if !skipValidate {
			for _, itemRef := range items {
				if err := validateNotGhost(svc, project, itemRef); err != nil {
					return err
				}
			}
		}

		// Duplicate guard: fetch existing links per item and filter.
		type workUnit struct {
			itemRef   string
			toLink    []api.JiraExternalItem
			skipped   []string // jira keys already linked
		}
		var plan []workUnit
		for _, itemRef := range items {
			existing, err := svc.Jira.GetLinks(project, itemRef)
			if err != nil {
				return fmt.Errorf("checking existing links on %s: %w", itemRef, err)
			}
			existingSet := map[string]bool{}
			for _, e := range existing {
				existingSet[strings.ToUpper(e.ExternalItemID)] = true
			}
			unit := workUnit{itemRef: itemRef}
			for _, ex := range templates {
				if existingSet[strings.ToUpper(ex.ExternalItemID)] {
					unit.skipped = append(unit.skipped, ex.ExternalItemID)
					continue
				}
				unit.toLink = append(unit.toLink, ex)
			}
			plan = append(plan, unit)
		}

		// If nothing to do, bail out cleanly.
		totalToLink := 0
		for _, u := range plan {
			totalToLink += len(u.toLink)
		}
		if totalToLink == 0 {
			fmt.Println("All requested Jira issues are already linked. Nothing to do.")
			for _, u := range plan {
				for _, k := range u.skipped {
					fmt.Printf("  already linked: %s -> %s (%s)\n", k, u.itemRef, project)
				}
			}
			return nil
		}

		// Show the plan.
		fmt.Printf("Will link (project=%s, reason=%q):\n", project, reason)
		for _, u := range plan {
			for _, ex := range u.toLink {
				fmt.Printf("  %s %q -> %s\n", ex.ExternalItemID, ex.ExternalItemTitle, u.itemRef)
			}
			for _, k := range u.skipped {
				fmt.Printf("  (skip) %s already linked to %s\n", k, u.itemRef)
			}
		}

		if !skipConfirm {
			if !confirm("Proceed?") {
				return fmt.Errorf("aborted")
			}
		}

		// Execute.
		for _, u := range plan {
			if len(u.toLink) == 0 {
				continue
			}
			if err := svc.Jira.CreateLinks(project, u.itemRef, u.toLink); err != nil {
				return fmt.Errorf("linking to %s: %w", u.itemRef, err)
			}
			fmt.Printf("Linked %d issue(s) to %s.\n", len(u.toLink), u.itemRef)
		}

		// Verify.
		fmt.Println("Verification:")
		for _, u := range plan {
			links, err := svc.Jira.GetLinks(project, u.itemRef)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  warning: could not verify %s: %v\n", u.itemRef, err)
				continue
			}
			fmt.Printf("  %s now has %d Jira link(s)\n", u.itemRef, len(links))
		}
		return nil
	},
}

var jiraRemoveCmd = &cobra.Command{
	Use:     "remove <item-ref> [<item-ref>...]",
	Aliases: []string{"unlink", "rm"},
	Short:   "Unlink one or more Jira issues from one or more Matrix items",
	Long: `Remove Jira-issue links from Matrix items. Requires 'mxreq jira init' to have been run.

The CLI fetches existing links first so you only need to pass --issue (the
title and URL needed for the DELETE body are read from the server).

If --issue is omitted, any item that has exactly one Jira link is auto-
unlinked. Items with multiple links will error unless you disambiguate with
--issue.

The --reason is required by convention but not sent to the wfgw endpoint.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireJiraConfigured(); err != nil {
			return err
		}
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}

		issuesFlag, _ := cmd.Flags().GetStringSlice("issue")
		pluginID, _ := cmd.Flags().GetInt("plugin-id")
		skipConfirm, _ := cmd.Flags().GetBool("yes")
		reason, _ := cmd.Flags().GetString("reason")

		items := make([]string, len(args))
		for i, a := range args {
			items[i] = upperRef(a)
		}

		type workUnit struct {
			itemRef  string
			toUnlink []api.JiraExternalItem
			missing  []string
			auto     bool // true when the target was picked because there's only one link
		}
		toExternal := func(itemRef string, hit api.JiraLinkedIssue) api.JiraExternalItem {
			plug := hit.Plugin
			if plug == 0 {
				plug = pluginID
			}
			return api.JiraExternalItem{
				ExternalItemID:    hit.ExternalItemID,
				ExternalItemTitle: hit.ExternalItemTitle,
				ExternalItemURL:   hit.ExternalItemURL,
				Plugin:            plug,
				MatrixItemIDs:     []string{itemRef},
			}
		}
		var plan []workUnit
		for _, itemRef := range items {
			existing, err := svc.Jira.GetLinks(project, itemRef)
			if err != nil {
				return fmt.Errorf("fetching links for %s: %w", itemRef, err)
			}
			unit := workUnit{itemRef: itemRef}

			if len(issuesFlag) == 0 {
				// Auto-pick: only valid when exactly one link exists.
				switch len(existing) {
				case 0:
					unit.missing = append(unit.missing, "(no links present)")
				case 1:
					unit.auto = true
					unit.toUnlink = append(unit.toUnlink, toExternal(itemRef, existing[0]))
				default:
					keys := make([]string, 0, len(existing))
					for _, e := range existing {
						keys = append(keys, e.ExternalItemID)
					}
					return fmt.Errorf(
						"%s has %d Jira links — pass --issue to pick one (found: %s)",
						itemRef, len(existing), strings.Join(keys, ", "),
					)
				}
				plan = append(plan, unit)
				continue
			}

			byKey := map[string]api.JiraLinkedIssue{}
			for _, e := range existing {
				byKey[strings.ToUpper(e.ExternalItemID)] = e
			}
			for _, k := range issuesFlag {
				hit, ok := byKey[strings.ToUpper(k)]
				if !ok {
					unit.missing = append(unit.missing, k)
					continue
				}
				unit.toUnlink = append(unit.toUnlink, toExternal(itemRef, hit))
			}
			plan = append(plan, unit)
		}

		totalToUnlink := 0
		for _, u := range plan {
			totalToUnlink += len(u.toUnlink)
		}
		if totalToUnlink == 0 {
			fmt.Println("No matching Jira links found on any of the specified items.")
			for _, u := range plan {
				for _, k := range u.missing {
					fmt.Printf("  not linked: %s on %s\n", k, u.itemRef)
				}
			}
			return fmt.Errorf("nothing to unlink")
		}

		fmt.Printf("Will unlink (project=%s, reason=%q):\n", project, reason)
		for _, u := range plan {
			suffix := ""
			if u.auto {
				suffix = " [auto-selected, only link on item]"
			}
			for _, ex := range u.toUnlink {
				fmt.Printf("  %s %q from %s%s\n", ex.ExternalItemID, ex.ExternalItemTitle, u.itemRef, suffix)
			}
			for _, k := range u.missing {
				fmt.Printf("  (skip) %s not linked to %s\n", k, u.itemRef)
			}
		}

		if !skipConfirm {
			if !confirm("Proceed?") {
				return fmt.Errorf("aborted")
			}
		}

		for _, u := range plan {
			if len(u.toUnlink) == 0 {
				continue
			}
			if err := svc.Jira.BreakLinks(project, u.itemRef, u.toUnlink, pluginID); err != nil {
				return fmt.Errorf("unlinking from %s: %w", u.itemRef, err)
			}
			fmt.Printf("Unlinked %d issue(s) from %s.\n", len(u.toUnlink), u.itemRef)
		}

		fmt.Println("Verification:")
		for _, u := range plan {
			links, err := svc.Jira.GetLinks(project, u.itemRef)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  warning: could not verify %s: %v\n", u.itemRef, err)
				continue
			}
			fmt.Printf("  %s now has %d Jira link(s)\n", u.itemRef, len(links))
		}
		return nil
	},
}

// validateNotGhost fetches the item and rejects it if it looks like a
// placeholder (empty title + version 0). Ghost items indicate the real item
// lives in a different project — the skill's Workflow B, step 2.
func validateNotGhost(svc *service.MatrixService, project, itemRef string) error {
	item, err := svc.Items.Get(project, itemRef, false)
	if err != nil {
		return fmt.Errorf("validating %s in %s: %w", itemRef, project, err)
	}
	if strings.TrimSpace(item.Title) == "" && item.MaxVersion == 0 {
		return fmt.Errorf(
			"ghost/placeholder item detected: %s in %s has empty title and v0 (folder %q) — the real item likely lives in a different project. Re-run against the correct project, or pass --skip-validate to proceed anyway",
			itemRef, project, item.FolderRef,
		)
	}
	return nil
}

// confirm prints a y/n prompt and returns true only for "y" or "yes".
func confirm(prompt string) bool {
	fmt.Printf("%s (y/n): ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes"
}

// decodeJiraMeta parses the externalMeta JSON string that wfgw returns as a
// blob inside each link.
func decodeJiraMeta(raw string) api.JiraExternalMeta {
	var meta api.JiraExternalMeta
	if raw == "" {
		return meta
	}
	_ = json.Unmarshal([]byte(raw), &meta)
	return meta
}

// shortDate returns the YYYY-MM-DD prefix of a datetime string.
func shortDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

// truncate clips a string to n runes with an ellipsis suffix.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}
