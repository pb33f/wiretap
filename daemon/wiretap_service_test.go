package daemon

// this test does not work. I have no idea how to test the platform. It will take a decent amount of work to dig deep.
// func TestWiretapService(t *testing.T) {
// 	config := `
// redirectURL: http://localhost:8000
// port: 8886
// monitorPort: 9093
// webSocketPort: 9094
// hardValidation: false
// mockModeList:
//   - "**/mock-this-path"
// strictRedirectLocation: true
// ignoreRedirects:
//   - "**"
// paths:
//   /i/am/rewriting/this/path*:
//     target: http://localhost:8886
//     rewriteId: "core-rewrite"
//     secure: false
//     headers:
//       drop:
//         - Origin
//         - Referer
//     pathRewrite:
//       "^/i/am/rewriting/this/path": ""

//   `

// 	var wcConfig *shared.WiretapConfiguration
// 	_ = yaml.Unmarshal([]byte(config), &wcConfig)

// 	spec := `
// openapi: 3.0.0
// info:
//   title: User Management API
//   version: 1.0.0

// paths:
//   /mock-this-path:
//     get:
//       summary: Get user details
//       responses:
//         '200':
//           description: User information
//           content:
//             application/json:
//               schema:
//                 $ref: '#/components/schemas/AdminResponse'
// components:
//   schemas:
//     UserResponse:
//       type: object
//       required:
//         - id
//         - email
//         - role
//       properties:
//         id:
//           type: string
//         email:
//           type: string
//           format: email
//         role:
//           type: string
//           enum: [user]
//   `

// 	d, _ := libopenapi.NewDocument([]byte(spec))
// m, _ := d.BuildV3Model()

// 	wcConfig.CompileVariables()
// 	wcConfig.CompileHardErrorList()
// 	wcConfig.CompilePaths()
// 	wcConfig.CompileMockModeList()

// 	ws := NewWiretapService(d, wcConfig)

// 	request, _ := http.NewRequest(http.MethodGet, "http://localhost:8000/i/am/rewriting/this/path/mock-this-path", bytes.NewReader([]byte(`data`)))
// 	request.Header.Set(helpers.ContentTypeHeader, "application/json")
// 	request.Header.Set(helpers.AuthorizationHeader, "admin:changeme")

// 	// request.Body

// 	id, _ := uuid.NewUUID()
// 	// create a new request that can be passed over to the service.
// 	requestModel := &model.Request{
// 		Id:          &id,
// 		HttpRequest: request,
// 	}

// 	ptermLog := &pterm.Logger{
// 		Formatter:  pterm.LogFormatterColorful,
// 		Writer:     os.Stdout,
// 		Level:      pterm.LogLevelDebug,
// 		ShowTime:   true,
// 		TimeFormat: "2006-01-02 15:04:05",
// 		MaxWidth:   180,
// 		KeyStyles: map[string]pterm.Style{
// 			"error":  *pterm.NewStyle(pterm.FgRed, pterm.Bold),
// 			"err":    *pterm.NewStyle(pterm.FgRed, pterm.Bold),
// 			"caller": *pterm.NewStyle(pterm.FgGray, pterm.Bold),
// 		},
// 	}

// 	handler := pterm.NewSlogHandler(ptermLog)
// 	wcConfig.Logger = slog.New(handler)

// 	storeManager := bus.GetBus().GetStoreManager()
// 	controlsStore := storeManager.CreateStoreWithType(controls.ControlServiceChan, reflect.TypeOf(wcConfig))
// 	controlsStore.Put(shared.ConfigKey, wcConfig, nil)

// 	// ws.rewritingPath(request, wcConfig)
// 	// ws.rewritingPath(requestModel.HttpRequest, wcConfig)
// 	// p, er, s := paths.FindPath(request, &m.Model)

// 	// create a new request that can be passed over to the service.
// 	ws.handleHttpRequest(requestModel)

// 	// ws.handleMockRequest(requestModel, wcConfig, request)

// }
