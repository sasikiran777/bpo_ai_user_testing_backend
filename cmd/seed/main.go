package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"ai_testing/internal/config"
	db "ai_testing/internal/db"
	testsmodel "ai_testing/internal/modules/tests/model"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func main() {
	cfg := config.Load()

	database, err := db.ConnectPostgres(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer database.Close()

	ctx := context.Background()

	english, err := ensureTest(ctx, database, testsmodel.Test{
		Name:        "English",
		Code:        "english",
		Description: "Measures real-world English communication with timed writing, comprehension, and speaking.",
		Instruction: []string{
			"You can take this test only once.",
			"Do not refresh/close the tab or switch away. Leaving the test may mark it as failed.",
			"Writing: 5 minutes. Reading: 5 minutes. Speaking: 3 minutes (auto-recording)",
			"Microphone access is required for the Speaking section.",
			"Your activity (tab changes / focus loss) is tracked silently during the test.",
		},
		IsActive: true,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if _, err := ensureTest(ctx, database, testsmodel.Test{
		Name:        "Agentic AI",
		Code:        "agentic_ai",
		Description: "Future module for agentic AI workflows and scenario-based testing.",

		IsActive: false,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	sections := []testsmodel.TestSectionMapping{
		{TestID: english.ID, Name: "Read", Description: "Comprehension and Reading", MaxMarks: 10, MaxTime: 5, IsActive: true},
		{TestID: english.ID, Name: "Write", Description: "Writing", MaxMarks: 10, MaxTime: 5, IsActive: true},
		{TestID: english.ID, Name: "Speak", Description: "Speaking", MaxMarks: 10, MaxTime: 3, IsActive: true},
	}

	var readSectionID uuid.UUID
	var speakSectionID uuid.UUID
	var hasRead bool
	var hasSpeak bool

	for _, s := range sections {
		sec, err := ensureTestSection(ctx, database, s)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if sec.Name == "Read" {
			readSectionID = sec.ID
			hasRead = true
		}
		if sec.Name == "Speak" {
			speakSectionID = sec.ID
			hasSpeak = true
		}
	}

	if !hasRead || !hasSpeak {
		fmt.Fprintln(os.Stderr, "missing required english sections")
		os.Exit(1)
	}

	if err := deactivateSpeakingTopics(ctx, database, speakSectionID); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	speakingTopics := []string{
		"Explain a skill you improved recently and how you practiced it.",
		"Describe a hobby you enjoy and why you like it.",
		"Describe a place in your city you like to visit.",
		"Describe a time you helped someone and how you felt about it.",
		"Describe a person who has influenced you positively.",
		"Describe a book you have read recently and what you learned from it.",
		"Describe a movie or series you watched recently and why you recommend it.",
		"Describe a memorable trip you took and what made it special.",
		"Describe an important event you attended (such as a wedding or festival).",
		"Describe a meal you enjoyed and what made it memorable.",
		"Describe a piece of technology you use often and how it helps you.",
		"Describe a difficult choice you made and what happened after.",
		"Describe a time you learned something new and why you decided to learn it.",
		"Describe a time you worked with a team and what role you played.",
		"Describe a goal you want to achieve in the next year and why.",
		"Describe a daily routine that helps you stay productive.",
		"Describe a time you were late and what you did about it.",
		"Describe a gift you gave someone and why you chose it.",
		"Describe a person you admire and what qualities you respect.",
		"Describe a change you would like to make in your life and why.",
	}

	for _, topic := range speakingTopics {
		if err := ensureSpeakingTopic(ctx, database, testsmodel.SpeakingTopic{
			TestSectionMappingID: speakSectionID,
			Topic:                topic,
			IsActive:             true,
		}); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	readingSets := []testsmodel.ReadingComprehension{
		{
			TestSectionMappingID: readSectionID,
			Title:                "A Small Change, Big Impact",
			Passage:              "Mina joined a support team that struggled with long response times. Instead of working harder, she mapped the workflow and noticed that many delays came from repeating the same explanations. She created a short library of reusable templates and added a checklist for common issues. Within two weeks, the team reduced average response time by 20%. Mina also asked teammates to suggest improvements, which helped them feel ownership of the process.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What was the main cause of delays?", "options": []string{"Too many meetings", "Repeated explanations", "Slow internet", "Lack of staff"}, "answer": "Repeated explanations"},
				{"type": "blank", "question": "The team reduced average response time by __%.", "answer": "20"},
				{"type": "short", "question": "Why did Mina ask teammates for suggestions?", "answer_key": "To build ownership and improve the process"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "Remote Work and Routine",
			Passage:              "When Ravi began working remotely, he enjoyed the flexibility but soon felt his days were blending together. To regain structure, he started his morning with a short plan: three priority tasks, one learning goal, and two short breaks. He also ended the day by writing a quick summary of what was completed and what could wait. This routine made it easier for him to disconnect after work and reduced stress.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What problem did Ravi face at first?", "options": []string{"He had too many calls", "His days felt unstructured", "He missed commuting", "He had no internet"}, "answer": "His days felt unstructured"},
				{"type": "blank", "question": "Ravi planned __ priority tasks each morning.", "answer": "three"},
				{"type": "short", "question": "How did the end-of-day summary help Ravi?", "answer_key": "Helped disconnect and reduce stress by clarifying what was done and what can wait"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "The Power of Clear Questions",
			Passage:              "During a meeting, a team argued about why a feature was failing. Some blamed the code, others blamed the requirements. A junior engineer asked one simple question: 'What exact behavior do we expect, and what do we actually see?' The team wrote both down and realized the requirement had changed after the last update. By clarifying expectations, they stopped guessing and fixed the issue quickly.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What helped the team resolve the argument?", "options": []string{"A longer meeting", "Clarifying expected vs actual behavior", "More people joined", "They stopped working"}, "answer": "Clarifying expected vs actual behavior"},
				{"type": "blank", "question": "The team realized the ______ had changed.", "answer": "requirement"},
				{"type": "short", "question": "What is the lesson of the passage?", "answer_key": "Clear questions reduce guessing and speed up problem-solving"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "Learning From Metrics",
			Passage:              "A product team tracked daily sign-ups and celebrated whenever the number rose. But sign-ups alone did not tell the whole story. When they added a metric for weekly active users, they found that many new users never returned after the first day. The team then improved onboarding by simplifying the first steps and adding clear tips. Over the next month, weekly active users increased steadily.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What did the new metric reveal?", "options": []string{"Users loved the product", "Many users did not return after day one", "Sign-ups were incorrect", "Marketing stopped working"}, "answer": "Many users did not return after day one"},
				{"type": "blank", "question": "They improved ______ by simplifying first steps.", "answer": "onboarding"},
				{"type": "short", "question": "Why were sign-ups alone not enough?", "answer_key": "They did not measure retention or ongoing usage"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "A Simple Communication Rule",
			Passage:              "Neha noticed that misunderstandings often started with vague messages like 'it’s not working' or 'please fix ASAP.' She introduced a simple rule: every issue report must include the context, the exact steps to reproduce, and the expected result. At first, some people resisted because it felt like extra work. But soon, problems were resolved faster because engineers had clearer information from the start.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What rule did Neha introduce?", "options": []string{"No messages after 6 PM", "Issue reports must include context and steps", "Only managers can report issues", "All issues must be urgent"}, "answer": "Issue reports must include context and steps"},
				{"type": "blank", "question": "Reports needed steps to ______ the issue.", "answer": "reproduce"},
				{"type": "short", "question": "Why did resolution become faster?", "answer_key": "Clearer, complete information reduced back-and-forth"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "Choosing the Right Trade-off",
			Passage:              "A startup wanted to launch a new feature quickly, but the engineering team warned that the design could create future maintenance costs. They agreed to a smaller first version: only the core use case, fewer settings, and simple reports. The product launched on time and users began giving feedback. Later, the team expanded the feature based on real usage instead of assumptions.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What approach did the team choose?", "options": []string{"Delay launch until perfect", "Launch a smaller first version", "Cancel the feature", "Add more settings"}, "answer": "Launch a smaller first version"},
				{"type": "blank", "question": "They expanded later based on real ______.", "answer": "usage"},
				{"type": "short", "question": "What benefit came from the smaller first version?", "answer_key": "Shipped on time and learned from real feedback"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "The Importance of Breaks",
			Passage:              "During exam season, Arjun studied for hours without stopping. He felt busy, but his memory was weak and he often reread the same pages. A mentor suggested a simple change: study for 25 minutes, then take a 5-minute break. Arjun followed the routine and found he could focus better and finish topics faster. The small breaks helped his brain recover and reduced fatigue.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What was Arjun’s problem?", "options": []string{"He studied too little", "His focus and memory were weak", "He took too many breaks", "He had no mentor"}, "answer": "His focus and memory were weak"},
				{"type": "blank", "question": "Arjun studied for __ minutes before each break.", "answer": "25"},
				{"type": "short", "question": "Why did breaks help?", "answer_key": "They improved focus and reduced fatigue"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "A Customer-Centered Update",
			Passage:              "A company released an app update that looked modern but removed a feature many users relied on. Support tickets increased and ratings dropped. Instead of arguing, the team reviewed the feedback and restored the feature with a better design. They also added a short message in the app explaining the change. Users appreciated the transparency and ratings improved over time.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What happened after the update?", "options": []string{"Tickets decreased", "Tickets increased and ratings dropped", "Users stopped using phones", "No one noticed"}, "answer": "Tickets increased and ratings dropped"},
				{"type": "blank", "question": "Users appreciated the team's ______.", "answer": "transparency"},
				{"type": "short", "question": "What did the team do to address the issue?", "answer_key": "Restored the feature and communicated the change"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "Planning for Risk",
			Passage:              "Before launching a marketing campaign, the team created a checklist of possible risks: broken links, wrong pricing, and slow landing pages. They assigned owners to test each item and set up monitoring to catch errors quickly. On launch day, one link failed, but it was detected within minutes and fixed. The campaign continued smoothly, and the team avoided major losses.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What helped detect the broken link quickly?", "options": []string{"A lucky user", "Monitoring and testing owners", "A competitor", "No one noticed"}, "answer": "Monitoring and testing owners"},
				{"type": "blank", "question": "They created a checklist of possible ______.", "answer": "risks"},
				{"type": "short", "question": "Why is assigning owners useful?", "answer_key": "Accountability ensures tasks are tested and completed"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "Consistency Builds Trust",
			Passage:              "Customers often judge a service by small details. One hotel trained staff to follow a consistent greeting, provide clear directions, and confirm requests in writing. These steps were simple but reduced confusion and complaints. Over time, guests began mentioning reliability in reviews. The hotel did not become perfect overnight, but consistent habits made the experience feel professional.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What reduced complaints?", "options": []string{"Higher prices", "Consistent staff habits", "Less staff training", "More advertising"}, "answer": "Consistent staff habits"},
				{"type": "blank", "question": "Guests mentioned ______ in reviews.", "answer": "reliability"},
				{"type": "short", "question": "What is the main idea?", "answer_key": "Small consistent actions improve customer trust"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "Reducing Meeting Time",
			Passage:              "A team complained that meetings took too long and rarely ended with decisions. They tried a new approach: every meeting required an agenda, a time limit, and a clear owner for next steps. If an agenda was missing, the meeting was cancelled. After a month, the team held fewer meetings and made faster decisions, because discussions stayed focused.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What rule was introduced?", "options": []string{"Meetings must be longer", "Meetings need an agenda and time limit", "Only executives can speak", "No decisions allowed"}, "answer": "Meetings need an agenda and time limit"},
				{"type": "blank", "question": "If an agenda was missing, the meeting was ______.", "answer": "cancelled"},
				{"type": "short", "question": "Why did decisions become faster?", "answer_key": "Focused discussions and clear next steps"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "The Value of Practice",
			Passage:              "Sara wanted to improve her public speaking, but she only practiced right before presentations. She decided to practice in smaller ways: speaking up in team meetings, recording short explanations, and asking for feedback. At first, she felt uncomfortable. Over time, her confidence increased, and she stopped rushing her words. Consistent practice helped her improve more than last-minute preparation.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What helped Sara improve most?", "options": []string{"Avoiding meetings", "Consistent small practice", "Speaking less", "Waiting for talent"}, "answer": "Consistent small practice"},
				{"type": "blank", "question": "Sara started ______ short explanations.", "answer": "recording"},
				{"type": "short", "question": "Why is last-minute preparation less effective?", "answer_key": "It does not build skills gradually or confidence"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "Clear Ownership",
			Passage:              "A bug stayed open for weeks because everyone assumed someone else was fixing it. Finally, the team introduced a rule: every issue must have one clear owner, even if multiple people contribute. The owner’s job was to coordinate, update progress, and close the issue. After this change, fewer tasks were forgotten, and work became more predictable.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "Why did the bug remain open?", "options": []string{"No one cared", "No clear owner", "Too many owners", "The bug was fake"}, "answer": "No clear owner"},
				{"type": "blank", "question": "Every issue must have one clear ______.", "answer": "owner"},
				{"type": "short", "question": "What is the owner’s job?", "answer_key": "Coordinate, communicate progress, and ensure completion"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "A Healthy Response to Failure",
			Passage:              "After a failed demo, the team felt embarrassed and wanted to blame each other. Their manager suggested a short review focused on learning: what went well, what didn’t, and what to change next time. The team realized their testing plan was incomplete and agreed on a checklist for future demos. The next demo went smoothly, and the team felt more prepared.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What did the manager suggest?", "options": []string{"Ignore the failure", "A learning-focused review", "Blame one person", "Cancel future demos"}, "answer": "A learning-focused review"},
				{"type": "blank", "question": "They agreed on a ______ for future demos.", "answer": "checklist"},
				{"type": "short", "question": "What did the team learn?", "answer_key": "Testing plan was incomplete and needed structure"},
			}),
			IsActive: true,
		},
		{
			TestSectionMappingID: readSectionID,
			Title:                "Keeping Promises",
			Passage:              "A freelancer gained clients by promising quick delivery, but sometimes missed deadlines. To improve, she began estimating time more carefully and setting expectations early. If a task was delayed, she informed the client immediately and offered options. Clients valued the honesty, and she noticed fewer stressful last-minute rushes. Reliability came from realistic planning, not optimism.",
			Questions: mustMarshalJSON([]map[string]any{
				{"type": "mcq", "question": "What improved the freelancer’s reliability?", "options": []string{"More optimism", "Realistic planning and communication", "Working all night", "Ignoring clients"}, "answer": "Realistic planning and communication"},
				{"type": "blank", "question": "Clients valued the ______.", "answer": "honesty"},
				{"type": "short", "question": "What is the main lesson?", "answer_key": "Reliability requires accurate estimates and communication"},
			}),
			IsActive: true,
		},
	}

	for _, r := range readingSets {
		if err := ensureReadingComprehension(ctx, database, r); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	fmt.Println("seed_complete")
}

func ensureTest(ctx context.Context, db *bun.DB, test testsmodel.Test) (*testsmodel.Test, error) {
	var existing testsmodel.Test
	if err := db.NewSelect().
		Model(&existing).
		Where("code = ?", test.Code).
		Limit(1).
		Scan(ctx); err == nil {
		return &existing, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if _, err := db.NewInsert().
		Model(&test).
		Returning("id").
		Exec(ctx, &test.ID); err != nil {
		return nil, err
	}

	return &test, nil
}

func ensureTestSection(ctx context.Context, db *bun.DB, s testsmodel.TestSectionMapping) (*testsmodel.TestSectionMapping, error) {
	var existing testsmodel.TestSectionMapping
	if err := db.NewSelect().
		Model(&existing).
		Where("test_id = ?", s.TestID).
		Where("name = ?", s.Name).
		Limit(1).
		Scan(ctx); err == nil {
		return &existing, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if _, err := db.NewInsert().
		Model(&s).
		Returning("id").
		Exec(ctx, &s.ID); err != nil {
		return nil, err
	}
	return &s, nil
}

func ensureSpeakingTopic(ctx context.Context, db *bun.DB, t testsmodel.SpeakingTopic) error {
	var existing testsmodel.SpeakingTopic
	if err := db.NewSelect().
		Model(&existing).
		Where("test_section_mapping_id = ?", t.TestSectionMappingID).
		Where("topic = ?", t.Topic).
		Limit(1).
		Scan(ctx); err == nil {
		if existing.IsActive {
			return nil
		}
		_, err := db.NewUpdate().
			Model(&existing).
			Set("is_active = true").
			WherePK().
			Exec(ctx)
		return err
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err := db.NewInsert().Model(&t).Exec(ctx)
	return err
}
func deactivateSpeakingTopics(ctx context.Context, db *bun.DB, sectionID uuid.UUID) error {
	_, err := db.NewUpdate().
		Model((*testsmodel.SpeakingTopic)(nil)).
		Set("is_active = false").
		Where("test_section_mapping_id = ?", sectionID).
		Exec(ctx)
	return err
}

func ensureReadingComprehension(ctx context.Context, db *bun.DB, r testsmodel.ReadingComprehension) error {
	var existing testsmodel.ReadingComprehension
	if err := db.NewSelect().
		Model(&existing).
		Where("test_section_mapping_id = ?", r.TestSectionMappingID).
		Where("title = ?", r.Title).
		Limit(1).
		Scan(ctx); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err := db.NewInsert().Model(&r).Exec(ctx)
	return err
}

func mustMarshalJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
