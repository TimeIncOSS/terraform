package pingdom

import (
	"bytes"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/russellcardullo/go-pingdom/pingdom"
)

func resourcePingdomHttpCheck() *schema.Resource {
	return &schema.Resource{
		Create: resourcePingdomHttpCheckCreate,
		Read:   resourcePingdomHttpCheckRead,
		Update: resourcePingdomHttpCheckUpdate,
		Delete: resourcePingdomHttpCheckDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Optional: true,
			},

			"interval_in_minutes": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  5,
			},

			"is_paused": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},

			"notify_when_down": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2,
			},

			"notify": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				// AllowedValues: []string{"android", "email", "iphone", "sms", "twitter"},
			},

			"repeat_notification_every": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},

			"notify_when_back_up": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  1,
			},
		},
	}
}

func resourcePingdomHttpCheckCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*pingdom.Client)

	check := pingdom.HttpCheck{
		Name:                   d.Get("name").(string),
		Hostname:               d.Get("hostname").(string),
		UseLegacyNotifications: true,
	}

	if v, ok := d.GetOk("interval_in_minutes"); ok {
		check.Resolution = v.(int)
	}
	if v, ok := d.GetOk("is_paused"); ok {
		check.Paused = v.(bool)
	}

	if v, ok := d.GetOk("notify"); ok {
		log.Printf("[DEBUG] Got notifications: %#v", v)
		notifications := expandCheckNotification(&check, v)
		check = *notifications
	}

	if v, ok := d.GetOk("notify_when_down"); ok {
		check.SendNotificationWhenDown = v.(int)
	}

	if v, ok := d.GetOk("repeat_notification_every"); ok {
		check.NotifyAgainEvery = v.(int)
	}

	if v, ok := d.GetOk("notify_when_back_up"); ok {
		check.NotifyWhenBackup = v.(bool)
	}

	log.Printf("[DEBUG] Creating Pingdom check: %#v", check)

	response, err := conn.Checks.Create(&check)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Pingdom check created: %#v", response)
	d.SetId(fmt.Sprintf("%d", response.ID))

	return resourcePingdomHttpCheckRead(d, meta)
}

func resourcePingdomHttpCheckRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*pingdom.Client)

	id, err := strconv.ParseInt(d.Id(), 0, 0)
	if err != nil {
		return fmt.Errorf("Error pasing ID: %#v", err)
	}
	check, err := conn.Checks.Read(int(id))
	if err != nil {
		return err
	}

	d.Set("name", check.Name)
	d.Set("hostname", check.Hostname)
	d.Set("is_paused", check.Paused)
	d.Set("interval_in_minutes", check.Resolution)

	// contactids	Contact identifiers. For example contactids=154325,465231,765871

	log.Printf("[DEBUG] Running flattenCheckNotifications(android = %#v)", check.SendToAndroid)
	d.Set("notify", flattenCheckNotifications(check))

	d.Set("notify_when_down", check.SendNotificationWhenDown)
	d.Set("repeat_notification_every", check.NotifyAgainEvery)
	d.Set("notify_when_back_up", check.NotifyWhenBackup)

	// tags	Check tags	Comma separated strings	no

	return nil
}

func resourcePingdomHttpCheckUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*pingdom.Client)

	id, err := strconv.ParseInt(d.Id(), 0, 0)
	if err != nil {
		return fmt.Errorf("Error pasing ID: %#v", err)
	}

	check := pingdom.HttpCheck{
		Name:                   d.Get("name").(string),
		Hostname:               d.Get("hostname").(string),
		UseLegacyNotifications: true,
	}

	if v, ok := d.GetOk("interval_in_minutes"); ok {
		check.Resolution = v.(int)
	}

	if v, ok := d.GetOk("is_paused"); ok {
		check.Paused = v.(bool)
	}

	notifications := expandCheckNotification(&check, d.Get("notify"))
	check = *notifications

	if v, ok := d.GetOk("notify_when_down"); ok {
		check.SendNotificationWhenDown = v.(int)
	}

	if v, ok := d.GetOk("repeat_notification_every"); ok {
		check.NotifyAgainEvery = v.(int)
	}

	if v, ok := d.GetOk("notify_when_back_up"); ok {
		check.NotifyWhenBackup = v.(bool)
	}

	log.Printf("[DEBUG] Updating Pingdom HTTP check: %#v", check)
	response, err := conn.Checks.Update(int(id), &check)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Pingdom HTTP check updated: %#v", response)

	return resourcePingdomHttpCheckRead(d, meta)
}

func resourcePingdomHttpCheckDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*pingdom.Client)

	id, err := strconv.ParseInt(d.Id(), 0, 0)
	if err != nil {
		return fmt.Errorf("Error passing ID: %#v", err)
	}

	log.Printf("[DEBUG] Deleting Pingdom HTTP check: %#v", id)
	response, err := conn.Checks.Delete(int(id))
	log.Printf("[DEBUG] Pingdom HTTP check deleted: %#v", response)

	return err
}

func resourcePingdomHttpCheckNotification(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	for k, v := range m {
		buf.WriteString(fmt.Sprintf("%s=%t-", k, v))
	}

	return hashcode.String(buf.String())
}

func expandCheckNotification(check *pingdom.HttpCheck, v interface{}) *pingdom.HttpCheck {
	notifications := v.([]interface{})
	log.Printf("[DEBUG] Received notifications: %#v", notifications)
	for _, value := range notifications {
		destination := value.(string)
		if destination == "android" {
			check.SendToAndroid = true
		}
		if destination == "email" {
			check.SendToEmail = true
		}
		if destination == "iphone" {
			check.SendToIPhone = true
		}
		if destination == "sms" {
			check.SendToSms = true
		}
		if destination == "twitter" {
			check.SendToTwitter = true
		}
	}
	return check
}

func flattenCheckNotifications(check *pingdom.CheckResponse) []string {
	notifications := []string{}

	log.Printf("[DEBUG] SendToAndroid => %#v", check.SendToAndroid)
	if check.SendToAndroid {
		notifications = append(notifications, "android")
	}
	if check.SendToEmail {
		notifications = append(notifications, "email")
	}
	if check.SendToIPhone {
		notifications = append(notifications, "iphone")
	}
	if check.SendToSms {
		notifications = append(notifications, "sms")
	}
	if check.SendToTwitter {
		notifications = append(notifications, "twitter")
	}

	return notifications
}
