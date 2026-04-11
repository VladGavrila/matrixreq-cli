package service

import "github.com/VladGavrila/matrixreq-cli/internal/client"

// MatrixService is the aggregate service providing access to all domain services.
type MatrixService struct {
	Client     *client.Client
	Projects   ProjectService
	Items      ItemService
	Categories CategoryService
	Fields     FieldService
	Users      UserService
	Groups     GroupService
	Search     SearchService
	Files      FileService
	Jobs       JobService
	Todos      TodoService
	Reports    ReportService
	Admin      AdminService
	Branches   BranchService
	Settings   SettingsService
	Jira       JiraService
}

// New creates a MatrixService with all domain services wired up.
func New(c *client.Client) *MatrixService {
	return &MatrixService{
		Client:     c,
		Projects:   &projectService{client: c},
		Items:      &itemService{client: c},
		Categories: &categoryService{client: c},
		Fields:     &fieldService{client: c},
		Users:      &userService{client: c},
		Groups:     &groupService{client: c},
		Search:     &searchService{client: c},
		Files:      &fileService{client: c},
		Jobs:       &jobService{client: c},
		Todos:      &todoService{client: c},
		Reports:    &reportService{client: c},
		Admin:      &adminService{client: c},
		Branches:   &branchService{client: c},
		Settings:   &settingsService{client: c},
		Jira:       &jiraService{client: c},
	}
}
