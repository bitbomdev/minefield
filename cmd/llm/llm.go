package custom

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/bitbomdev/minefield/cmd/helpers"
	apiv1 "github.com/bitbomdev/minefield/gen/api/v1"
	"github.com/bitbomdev/minefield/gen/api/v1/apiv1connect"
	"github.com/olekukonko/tablewriter"
	chromadb "github.com/philippgille/chromem-go"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

const PROMPT_TEMPLATE = `You are an AI assistant that helps users understand and work with a DSL (Domain Specific Language) for querying a graph database of supply chain security artifacts. You have access to documentation and examples about this DSL through the provided context.

If the user asks for a DSL query, convert their natural language into the appropriate DSL script. The DSL uses keywords like: dependencies, dependents, library, vuln, xor, or, and.

If the user asks general questions about the DSL or how it works, provide helpful explanations based on the context.

YOU CAN ONLY OUTPUT THE DSL QUERY. NO REGULAR LANGUAGE.

You cannot output periods, commas, or other punctuation.

If an '@' is used in a package name, even if a version is not included, leave it in and do not remove it.

If a user asks for vulnerablities for a package they mean dependencies of type vuln, for a query, and if we want to know what a vuln affects, dependents of type library.

Globsearch queries can be used to find anything that might exist in a node, not only names of nodes, so for types of packages, if it is inside of a purl you can find it, versions, ecosystem, etc. For vulns you can do globsearches like '*GHSA*', '*CVE*', etc, depending on what the user asks.

Try to surrond a globsearch pattern with as much glob as you can, be as general as possible.


If this is a leaderboard query, you should prefix your answer with 'leaderboard:'.
If this is a regular query, you should prefix your answer with 'query:'.
If this is a globsearch query, you should prefix your answer with 'globsearch:'.


Context information:

%s

---

User question: %s

Please provide a helpful response. If the question requires a DSL query, format it clearly as a DSL script.`

// options holds the command-line options.
type options struct {
	maxOutput          int
	showInfo           bool
	saveQuery          string
	addr               string
	output             string
	queryServiceClient apiv1connect.QueryServiceClient
	leaderboardServiceClient apiv1connect.LeaderboardServiceClient
	graphServiceClient apiv1connect.GraphServiceClient
	vectorDBPath       string
}

// AddFlags adds command-line flags to the provided cobra command.
func (o *options) AddFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&o.maxOutput, "max-output", 10, "maximum number of results to display")
	cmd.Flags().BoolVar(&o.showInfo, "show-info", true, "display the info column")
	cmd.Flags().StringVar(&o.addr, "addr", "http://localhost:8089", "address of the minefield server")
	cmd.Flags().StringVar(&o.vectorDBPath, "vector-db-path", "./db", "Path to the vector database")
	cmd.Flags().StringVar(&o.output, "output", "table", "output format (table or json)")
}

// Run executes the custom command with the provided arguments.
func (o *options) Run(cmd *cobra.Command, args []string) error {
	
	if os.Getenv("OPENAI_API_KEY") == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}
	db, err := chromadb.NewPersistentDB(o.vectorDBPath, false)
	if err != nil {
		return fmt.Errorf("failed to initialize ChromaDB: %w", err)
	}

	c := db.GetCollection("knowledge-base", nil)
	if err != nil {
		return fmt.Errorf("failed to get collection from ChromaDB: %w", err)
	}

	// Initialize chat messages
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: PROMPT_TEMPLATE,
		},
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	// Initialize client if not injected (for testing)
	if o.queryServiceClient == nil {
		o.queryServiceClient = apiv1connect.NewQueryServiceClient(
			http.DefaultClient,
			o.addr,
			connect.WithGRPC(),
			connect.WithSendGzip(),
		)
	}

	if o.leaderboardServiceClient == nil {
		o.leaderboardServiceClient = apiv1connect.NewLeaderboardServiceClient(
			http.DefaultClient,
			o.addr,
		)
	}

	if o.graphServiceClient == nil {
		o.graphServiceClient = apiv1connect.NewGraphServiceClient(
			http.DefaultClient,
			o.addr,
		)
	}

	fmt.Println("Starting chat session. Type 'exit' to end.")
	
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nYou: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		input = strings.TrimSpace(input)
		if strings.ToLower(input) == "exit" {
			fmt.Println("Ending chat session. Goodbye!")
			return nil
		}

		// Get context from ChromaDB query
		resultEmbeddings, err := c.Query(context.Background(), input, 13, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to query ChromaDB: %w", err)
		}

		// Build context text from results
		var contextText string
		for i := 0; i < 13 && i < len(resultEmbeddings); i++ {
			contextText += fmt.Sprintf("%s\n\n", resultEmbeddings[i].Content)
		}

		// Add user's message
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf(PROMPT_TEMPLATE, contextText, input),
		})

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT4,
				Messages: messages,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to create chat completion: %w", err)
		}

		script := resp.Choices[0].Message.Content

		// Execute the query and capture output
		var queryResult string
		if strings.TrimSpace(script) != "" {
			var buf strings.Builder

			// Check if this is a leaderboard query
			if strings.HasPrefix(strings.TrimSpace(script), "leaderboard:") {


				// Remove the "leaderboard:" prefix
				cleanScript := strings.TrimPrefix(strings.TrimSpace(script), "leaderboard:")
				fmt.Printf("\nAssistant: I'll help you with that. I'm going to use this leaderboard query:\n\"%s\"\n", cleanScript)

				req := connect.NewRequest(&apiv1.CustomLeaderboardRequest{Script: cleanScript})
				res, err := o.leaderboardServiceClient.CustomLeaderboard(cmd.Context(), req)
				
				if err != nil {
					queryResult = fmt.Sprintf("Leaderboard query failed: %v", err)
				} else if len(res.Msg.Queries) == 0 {
					queryResult = "No results found"
				} else {
					switch o.output {
					case "json":
						jsonOutput, err := helpers.FormatCustomQueriesJSON(res.Msg.Queries)
						if err != nil {
							queryResult = fmt.Sprintf("Failed to format JSON: %v", err)
						} else {
							queryResult = string(jsonOutput)
						}
					case "table":
						err = formatLeaderboardTable(&buf, res.Msg.Queries, o.maxOutput, o.showInfo)
						if err != nil {
							queryResult = fmt.Sprintf("Failed to format table: %v", err)
						} else {
							queryResult = buf.String()
						}
					}
				}
			} else if strings.HasPrefix(strings.TrimSpace(script), "query:") {
				// Remove the "query:" prefix
				cleanScript := strings.TrimPrefix(strings.TrimSpace(script), "query:")
				fmt.Printf("\nAssistant: I'll help you with that. I'm going to use this query:\n\"%s\"\n", cleanScript)

				req := connect.NewRequest(&apiv1.QueryRequest{Script: cleanScript})
				res, err := o.queryServiceClient.Query(cmd.Context(), req)
				
				if err != nil {
					queryResult = fmt.Sprintf("Query failed: %v", err)
				} else if len(res.Msg.Nodes) == 0 {
					queryResult = "No results found"
				} else {
					switch o.output {
					case "json":
						jsonOutput, err := helpers.FormatNodeJSON(res.Msg.Nodes)
						if err != nil {
							queryResult = fmt.Sprintf("Failed to format JSON: %v", err)
						} else {
							queryResult = string(jsonOutput)
						}
					case "table":
						err = formatTable(&buf, res.Msg.Nodes, o.maxOutput, o.showInfo)
						if err != nil {
							queryResult = fmt.Sprintf("Failed to format table: %v", err)
						} else {
							queryResult = buf.String()
						}
					}
				}
			} else if strings.HasPrefix(strings.TrimSpace(script), "globsearch:") {
				// Remove the "globsearch:" prefix
				pattern := strings.TrimPrefix(strings.TrimSpace(script), "globsearch:")
				fmt.Printf("\nAssistant: I'll help you with that. I'm going to use this pattern:\n\"%s\"\n", pattern)

				req := connect.NewRequest(&apiv1.GetNodesByGlobRequest{Pattern: pattern})
				res, err := o.graphServiceClient.GetNodesByGlob(cmd.Context(), req)
				
				if err != nil {
					queryResult = fmt.Sprintf("Query failed: %v", err)
				} else if len(res.Msg.Nodes) == 0 {
					queryResult = "No results found"
				} else {
					switch o.output {
					case "json":
						jsonOutput, err := helpers.FormatNodeJSON(res.Msg.Nodes)
						if err != nil {
							queryResult = fmt.Sprintf("Failed to format JSON: %v", err)
						} else {
							queryResult = string(jsonOutput)
						}
					case "table":
						err = formatTableGlobSearch(&buf, res.Msg.Nodes, o.maxOutput, o.showInfo)
						if err != nil {
							queryResult = fmt.Sprintf("Failed to format table: %v", err)
						} else {
							queryResult = buf.String()
						}
					}
				}
			} else {
				queryResult = "Sorry the query failed, please try again."
			}
		}

		// Add results to chat history
		feedbackMsg := fmt.Sprintf("Here are the results of the query:\n%s\nWhat else would you like to know?", queryResult)
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: feedbackMsg,
		})

		fmt.Println(feedbackMsg)
	}
}

// formatTable formats the nodes into a table and writes it to the provided writer.
func formatTable(w io.Writer, nodes []*apiv1.Node, maxOutput int, showInfo bool) error {
	table := tablewriter.NewWriter(w)
	headers := []string{"Name", "Type", "ID"}
	if showInfo {
		headers = append(headers, "Info")
	}
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	count := 0
	for _, node := range nodes {
		if count >= maxOutput {
			break
		}

		row := []string{
			node.Name,
			node.Type,
			strconv.FormatUint(uint64(node.Id), 10),
		}

		if showInfo {
			additionalInfo := helpers.ComputeAdditionalInfo(node)
			row = append(row, additionalInfo)
		}

		table.Append(row)
		count++
	}

	table.Render()
	return nil
}

// Add new function to format leaderboard table
func formatLeaderboardTable(w io.Writer, queries []*apiv1.Query, maxOutput int, showInfo bool) error {
	table := tablewriter.NewWriter(w)
	headers := []string{"Name", "Type", "ID", "Output"}
	if showInfo {
		headers = append(headers, "Info")
	}
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	count := 0
	for _, query := range queries {
		if count >= maxOutput {
			break
		}

		row := []string{
			query.Node.Name,
			query.Node.Type,
			strconv.FormatUint(uint64(query.Node.Id), 10),
			fmt.Sprint(len(query.Output)),
		}

		if showInfo {
			additionalInfo := helpers.ComputeAdditionalInfo(query.Node)
			row = append(row, additionalInfo)
		}

		table.Append(row)
		count++
	}

	table.Render()
	return nil
}

func formatTableGlobSearch(w io.Writer, nodes []*apiv1.Node, maxOutput int, showInfo bool) error {
    table := tablewriter.NewWriter(w)
    table.SetHeader([]string{"Name", "Type", "ID"})
    table.SetAutoWrapText(false)
    table.SetAutoFormatHeaders(true)

    for i, node := range nodes {
        if i >= maxOutput {
            break
        }
        table.Append([]string{
            node.Name,
            node.Type,
            strconv.FormatUint(uint64(node.Id), 10),
        })
    }

    table.Render()
    return nil
}

// New creates and returns a new Cobra command for executing custom query scripts.
func New() *cobra.Command {
	o := &options{}

	cmd := &cobra.Command{
		Use:               "llm [query]",
		Short:             "Create a chat session with an LLM to query the graph for leaderboards, queries, and globsearches",
		Long:              "Creates an LLM chat session to query the graph for leaderboards, queries, and globsearches, to end the chat session use the type 'exit'. This does use OpenAI, so you need to have the OPENAI_API_KEY environment variable set.",
		Args:              cobra.NoArgs,
		RunE:              o.Run,
		DisableAutoGenTag: true,
	}

	o.AddFlags(cmd)

	return cmd
}
