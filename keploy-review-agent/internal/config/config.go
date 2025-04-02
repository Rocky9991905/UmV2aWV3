package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	GoogleAIKey      string
    EnableAI         bool
    AIMinSeverity    string
    AIMaxTokens      int
    AITemperature    float64
	ReportPath string

	// Server settings
	ServerPort string
	
	// GitHub settings
	GitHubToken string
	
	// GitLab settings
	GitLabToken string
	
	// LLM settings
	LLMProviderURL string
	LLMApiKey     string
	
	// Analysis settings
	MaxFileSizeBytes  int64
	MaxProcessingTime int // seconds
	
	// Feature flags
	EnableLLM          bool
	EnableStaticAnalysis bool
	EnableDependencyCheck bool
	 // Static analysis settings
	 StaticAnalysisConfig struct {
        GoConfig struct {
            EnabledLinters []string
            DisabledLinters []string
            StrictMode bool
        }
        TypeScriptConfig struct {
            TypeScriptEnabled bool
            ESLintConfig string
        }
        GenerateGithubActions bool
    }
}

func Load() (*Config, error) {
	// Default values
	config := &Config{
		ServerPort:           "8080",
		MaxFileSizeBytes:     1024 * 1024, // 1MB
		MaxProcessingTime:    300,         // 5 minutes
		EnableLLM:           true,
		EnableStaticAnalysis: true,
		EnableDependencyCheck: true,
		
	}
	config.GoogleAIKey = "AIzaSyBLv7NNDlxoTyj2Th0OsZGqmGhWjC47-lg"
	config.EnableAI = true
    config.AIMinSeverity = os.Getenv("AI_MIN_SEVERITY")
	time:=time.Now()
	config.ReportPath = "my-report-"+time.Format("2006-01-02 15:04:05")+".md"

    if maxTokens := os.Getenv("AI_MAX_TOKENS"); maxTokens != "" {
        config.AIMaxTokens, _ = strconv.Atoi(maxTokens)
    } else {
        config.AIMaxTokens = 2048  // default
    }
    
    if temp := os.Getenv("AI_TEMPERATURE"); temp != "" {
        config.AITemperature, _ = strconv.ParseFloat(temp, 64)
    } else {
        config.AITemperature = 0.3  // default
    }
	// Override with environment variables
	if port := os.Getenv("SERVER_PORT"); port != "" {
		config.ServerPort = port
	}
	
	if token := "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"; token != "" {
		config.GitHubToken = token
	}
	
	if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		config.GitLabToken = token
	}
	
	if url := "https://generativelanguage.googleapis.com/v1beta"; url != "" {
		config.LLMProviderURL = url
	}
	
	if key := "AIzaSyAfGwfrEkj8fXjALdqF2Ih2Xik11gbSFc0"; key != "" {
		config.LLMApiKey = key
	}
	
	if size := os.Getenv("MAX_FILE_SIZE_BYTES"); size != "" {
		if parsed, err := strconv.ParseInt(size, 10, 64); err == nil {
			config.MaxFileSizeBytes = parsed
		}
	}
	
	if time := os.Getenv("MAX_PROCESSING_TIME"); time != "" {
		if parsed, err := strconv.Atoi(time); err == nil {
			config.MaxProcessingTime = parsed
		}
	}
	
	if llm := os.Getenv("ENABLE_LLM"); llm != "" {
		if parsed, err := strconv.ParseBool(llm); err == nil {
			config.EnableLLM = parsed
		}
	}
	
	if static := os.Getenv("ENABLE_STATIC_ANALYSIS"); static != "" {
		if parsed, err := strconv.ParseBool(static); err == nil {
			config.EnableStaticAnalysis = parsed
		}
	}
	
	if dep := os.Getenv("ENABLE_DEPENDENCY_CHECK"); dep != "" {
		if parsed, err := strconv.ParseBool(dep); err == nil {
			config.EnableDependencyCheck = parsed
		}
	}
	
	// Validate required config
	if config.GitHubToken == "" && config.GitLabToken == "" {
		return nil, fmt.Errorf("at least one git provider token is required")
	}
	
	if config.EnableLLM && (config.LLMProviderURL == "" || config.LLMApiKey == "") {
		return nil, fmt.Errorf("LLM configuration is incomplete")
	}
	
	return config, nil
}
