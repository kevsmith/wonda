# Multi-Agent Simulation System Overview

## Core System Architecture

A high-level design for a multi-agent simulation system targeted at fiction writers, designed to simulate single scenes within larger works of fiction.

## Design Principles

**Creative Tool, Not Content Generator**
The system treats the chronicle as raw creative ore for writers to refine, not as finished content. The goal is to generate rich source material that inspires writers, providing them with character insights, emotional beats, dialogue patterns, and narrative possibilities to explore and transform into their own work.

### Nomenclature

- **Scenario**: A template for a simulation that defines goals and optionally assigns specific goals to agents
- **Goal**: A condition which must be satisfied by the conclusion of the simulation (priority 1-5, with 1 being most important)
- **Character**: The definition of a role included in a scenario (e.g., "Scrappy youth from wrong side of town")
- **Agent**: An anonymous actor which takes on the role of a character during simulation
- **Simulation**: An executable instance of a scenario

### Primary Components

#### 1. Simulation Framework
- Orchestrates scenario execution
- Manages simulation lifecycle (initialization → running → completion)
- Coordinates agent turns/action scheduling
- Monitors overall goal completion and termination conditions
- Provides scene narration and event logging for the writer

#### 2. Agent Manager
- Instantiates agents from character definitions
- Assigns character-specific and shared goals
- Manages agent lifecycle and state
- Handles agent-to-agent communication protocols
- Coordinates LLM calls for agent decision-making

#### 3. Environment State System
- Maintains current scene state (location, objects, conditions)
- Tracks agent positions and relationships
- Manages environmental changes from agent actions
- Provides consistent world state for all agent queries
- Supports temporal progression (time of day, weather changes, etc.)

#### 4. MCP Server Interface Layer
The simulation environment exposes capabilities through MCP servers that agents can use:
- **Perception Server**: `observe_scene()`, `listen()`, `sense_mood()`, `detect_objects()`
- **Action Server**: `move_to()`, `interact_with()`, `speak()`, `perform_action()`
- **Query Server**: `check_relationship()`, `recall_fact()`, `assess_situation()`
- **Communication Server**: `send_message()`, `broadcast()`, `whisper_to()`

#### 5. RAG Memory System
- Each agent maintains both short-term (context) and long-term (RAG) memory
- Stores: observations, interactions, learned facts, emotional states
- Retrieval based on relevance to current situation
- Memory decay/reinforcement based on importance and recency
- Shared memory pool for "common knowledge" accessible to all agents

#### 6. Goal Monitoring System
- Tracks progress toward individual and collective goals
- Evaluates goal completion conditions continuously
- Manages goal priorities and conflicts
- Triggers scenario completion when sufficient goals are met
- Provides feedback to agents about goal proximity

### Integration Points

- **LLM Provider Interface**: OpenAI-compatible API support (including Ollama, LM Studio, etc.)
- **MCP Protocol**: Standard Model Context Protocol for agent-environment interaction
- **RAG Storage**: Vector database for long-term memory storage and retrieval