package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jmpsec/osctrl/tags"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
)

// Helper function to convert a slice of tags into the data expected for output
func tagsToData(tgs []tags.AdminTag, header []string) [][]string {
	var data [][]string
	if header != nil {
		data = append(data, header)
	}
	for _, n := range tgs {
		data = append(data, tagToData(n, nil)...)
	}
	return data
}

// Helper function to convert a tag into the data expected for output
func tagToData(t tags.AdminTag, header []string) [][]string {
	var data [][]string
	if header != nil {
		data = append(data, header)
	}
	_t := []string{
		t.CreatedAt.String(),
		t.Name,
		t.Description,
		t.Color,
		t.Icon,
		//t.EnvironmentID,
		t.CreatedBy,
		stringifyBool(t.AutoTag),
		tags.TagTypeDecorator(t.TagType),
	}
	data = append(data, _t)
	return data
}

func addTag(c *cli.Context) error {
	// Get values from flags
	env := c.String("env-uuid")
	if env == "" {
		fmt.Println("❌ environment is required")
		os.Exit(1)
	}
	name := c.String("name")
	if name == "" {
		fmt.Println("❌ tag name is required")
		os.Exit(1)
	}
	color := c.String("color")
	if color == "" {
		color = tags.RandomColor()
	}
	icon := c.String("icon")
	if icon == "" {
		icon = tags.DefaultTagIcon
	}
	description := c.String("description")
	tagType := c.String("tag-type")
	if dbFlag {
		e, err := envs.Get(env)
		if err != nil {
			return fmt.Errorf("❌ error env get - %s", err)
		}
		// TODO - Use the correct user
		if err := tagsmgr.NewTag(name, description, color, icon, appName, e.ID, false, tags.TagTypeParser(tagType)); err != nil {
			return fmt.Errorf("❌ %s", err)
		}
	} else if apiFlag {
		_, err := osctrlAPI.AddTag(env, name, color, icon, description, tags.TagTypeParser(tagType))
		if err != nil {
			return fmt.Errorf("❌ %s", err)
		}
	}
	if !silentFlag {
		fmt.Printf("✅ tag %s updated successfully\n", name)
	}
	return nil
}

func deleteTag(c *cli.Context) error {
	// Get values from flags
	env := c.String("env-uuid")
	if env == "" {
		fmt.Println("❌ environment is required")
		os.Exit(1)
	}
	name := c.String("name")
	if name == "" {
		fmt.Println("❌ tag name is required")
		os.Exit(1)
	}
	if dbFlag {
		e, err := envs.Get(env)
		if err != nil {
			return fmt.Errorf("❌ error env get - %s", err)
		}
		if err := tagsmgr.DeleteGet(name, e.ID); err != nil {
			return fmt.Errorf("❌ %s", err)
		}
	} else if apiFlag {
		_, err := osctrlAPI.DeleteTag(env, name)
		if err != nil {
			return fmt.Errorf("❌ %s", err)
		}
	}
	if !silentFlag {
		fmt.Printf("✅ tag %s deleted successfully\n", name)
	}
	return nil
}

func editTag(c *cli.Context) error {
	// Get values from flags
	env := c.String("env-uuid")
	if env == "" {
		fmt.Println("❌ environment is required")
		os.Exit(1)
	}
	name := c.String("name")
	if name == "" {
		fmt.Println("❌ tag name is required")
		os.Exit(1)
	}
	color := c.String("color")
	icon := c.String("icon")
	description := c.String("description")
	tagType := c.String("tag-type")
	if dbFlag {
		e, err := envs.Get(env)
		if err != nil {
			return fmt.Errorf("❌ error env get - %s", err)
		}
		t, err := tagsmgr.Get(name, e.ID)
		if err != nil {
			return fmt.Errorf("❌ %s", err)
		}
		if description != "" && description != t.Description {
			if err := tagsmgr.ChangeDescription(&t, description); err != nil {
				return fmt.Errorf("❌ %s", err)
			}
		}
		if color != "" && color != t.Color {
			if err := tagsmgr.ChangeColor(&t, color); err != nil {
				return fmt.Errorf("❌ %s", err)
			}
		}
		if icon != "" && icon != t.Icon {
			if err := tagsmgr.ChangeIcon(&t, icon); err != nil {
				return fmt.Errorf("❌ %s", err)
			}
		}
		if tagType != "" && tagType != tags.TagTypeDecorator(t.TagType) {
			if err := tagsmgr.ChangeTagType(&t, tags.TagTypeParser(tagType)); err != nil {
				return fmt.Errorf("❌ %s", err)
			}
		}
	} else if apiFlag {
		t, err := osctrlAPI.GetTag(env, name)
		if err != nil {
			return fmt.Errorf("❌ %s", err)
		}
		tt := t.TagType
		if tagType != "" && strings.ToUpper(tagType) != strings.ToUpper(tags.TagTypeDecorator(t.TagType)) {
			tt = tags.TagTypeParser(tagType)
		}
		_, err = osctrlAPI.EditTag(env, name, color, icon, description, tt)
		if err != nil {
			return fmt.Errorf("❌ %s", err)
		}
	}
	if !silentFlag {
		fmt.Printf("✅ tag %s updated successfully\n", name)
	}
	return nil
}

func showTag(c *cli.Context) error {
	// Get values from flags
	env := c.String("env-uuid")
	if env == "" {
		fmt.Println("❌ environment is required")
		os.Exit(1)
	}
	name := c.String("name")
	if name == "" {
		fmt.Println("❌ tag name is required")
		os.Exit(1)
	}
	var t tags.AdminTag
	if dbFlag {
		e, err := envs.Get(env)
		if err != nil {
			return fmt.Errorf("❌ error env get - %s", err)
		}
		t, err = tagsmgr.Get(name, e.ID)
		if err != nil {
			return fmt.Errorf("❌ %s", err)
		}
	} else if apiFlag {
		t, err = osctrlAPI.GetTag(env, name)
		if err != nil {
			return fmt.Errorf("❌ %s", err)
		}
	}
	fmt.Printf("Tag: %s\n", t.Name)
	fmt.Printf("Description: %s\n", t.Description)
	fmt.Printf("Color: %s\n", t.Color)
	fmt.Printf("Icon: %s\n", t.Icon)
	fmt.Printf("Created: %s\n", t.CreatedAt.String())
	fmt.Printf("CreatedBy: %s\n", t.CreatedBy)
	fmt.Printf("AutoTag: %s\n", stringifyBool(t.AutoTag))
	fmt.Printf("TagType: %s\n", tags.TagTypeDecorator(t.TagType))
	fmt.Println()
	return nil
}

func helperListTags(tgs []tags.AdminTag) error {
	header := []string{
		"Created",
		"Name",
		"Description",
		"Color",
		"Icon",
		"CreatedBy",
		"AutoTag",
		"TagType",
	}
	// Prepare output
	if formatFlag == jsonFormat {
		jsonRaw, err := json.Marshal(tgs)
		if err != nil {
			return fmt.Errorf("error marshaling - %s", err)
		}
		fmt.Println(string(jsonRaw))
	} else if formatFlag == csvFormat {
		data := tagsToData(tgs, header)
		w := csv.NewWriter(os.Stdout)
		if err := w.WriteAll(data); err != nil {
			return fmt.Errorf("error writting csv - %s", err)
		}
	} else if formatFlag == prettyFormat {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader(header)
		if len(tgs) > 0 {
			fmt.Printf("Existing tags (%d):\n", len(tgs))
			data := tagsToData(tgs, nil)
			table.AppendBulk(data)
		} else {
			fmt.Println("No tags")
		}
		table.Render()
	}
	return nil
}

func listTagsByEnv(c *cli.Context) error {
	// Get values from flags
	env := c.String("env-uuid")
	if env == "" {
		fmt.Println("❌ environment is required")
		os.Exit(1)
	}
	// Retrieve data
	var tgs []tags.AdminTag
	if dbFlag {
		e, err := envs.Get(env)
		if err != nil {
			return fmt.Errorf("❌ error env get - %s", err)
		}
		tgs, err = tagsmgr.GetByEnv(e.ID)
		if err != nil {
			return fmt.Errorf("❌ %s", err)
		}
	} else if apiFlag {
		tgs, err = osctrlAPI.GetTags(env)
		if err != nil {
			return fmt.Errorf("❌ %s", err)
		}
	}
	if err := helperListTags(tgs); err != nil {
		return fmt.Errorf("❌ %s", err)
	}
	return nil
}

func listAllTags(c *cli.Context) error {
	var tgs []tags.AdminTag
	if dbFlag {
		tgs, err = tagsmgr.All()
		if err != nil {
			return fmt.Errorf("❌ %s", err)
		}
	} else if apiFlag {
		tgs, err = osctrlAPI.GetAllTags()
		if err != nil {
			return fmt.Errorf("❌ %s", err)
		}
	}
	if err := helperListTags(tgs); err != nil {
		return fmt.Errorf("❌ %s", err)
	}
	return nil
}
