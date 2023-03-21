package brightbox

import (
	"context"
	"errors"
	"log"

	brightbox "github.com/brightbox/gobrightbox/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBrightboxCreate[O, I any](
	creator func(*brightbox.Client, context.Context, I) (*O, error),
	objectName string,
	updater func(*schema.ResourceData, *I) diag.Diagnostics,
	setter func(*schema.ResourceData, *O) diag.Diagnostics,
) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		client := meta.(*CompositeClient).APIClient

		log.Printf("[INFO] Creating %s", objectName)
		var objectOptions I
		errs := updater(d, &objectOptions)
		if errs.HasError() {
			return errs
		}
		log.Printf("[INFO] %s create configuration: %+v", objectName, objectOptions)
		object, err := creator(client, ctx, objectOptions)
		if err != nil {
			return diag.FromErr(err)
		}
		log.Printf("[DEBUG] setting details from returned object: %+v", *object)
		return setter(d, object)
	}
}

func resourceBrightboxRead[O any](
	reader func(*brightbox.Client, context.Context, string) (*O, error),
	objectName string,
	setter func(*schema.ResourceData, *O) diag.Diagnostics,
) schema.ReadContextFunc {
	return resourceBrightboxReadStatus(
		reader,
		objectName,
		setter,
		func(_ *O) bool {
			return false
		},
	)
}

func resourceBrightboxReadStatus[O any](
	reader func(*brightbox.Client, context.Context, string) (*O, error),
	objectName string,
	setter func(*schema.ResourceData, *O) diag.Diagnostics,
	missing func(*O) bool,
) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		client := meta.(*CompositeClient).APIClient

		log.Printf("[DEBUG] %s resource read called for %s", objectName, d.Id())

		object, err := reader(client, ctx, d.Id())
		if err != nil {
			var apierror *brightbox.APIError
			if errors.As(err, &apierror) {
				if apierror.StatusCode == 404 {
					log.Printf("[WARN] %s not found, removing from state: %s", objectName, d.Id())
					d.SetId("")
					return nil
				}
			}
			return diag.FromErr(err)
		}
		if missing(object) {
			log.Printf("[WARN] %s not found, removing from state: %s", objectName, d.Id())
			d.SetId("")
			return nil
		}

		log.Printf("[DEBUG] setting details from returned object: %+v", *object)
		return setter(d, object)
	}
}

func datasourceBrightboxRead[O any](
	reader func(*brightbox.Client, context.Context) ([]O, error),
	objectName string,
	setter func(*schema.ResourceData, *O) diag.Diagnostics,
	finderGenerator func(*schema.ResourceData) (func(O) bool, diag.Diagnostics),
) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		client := meta.(*CompositeClient).APIClient

		log.Printf("[DEBUG] %s data read called. Retrieving object list", objectName)

		objects, err := reader(client, ctx)
		if err != nil {
			return diag.FromErr(err)
		}

		findFunc, errs := finderGenerator(d)
		if errs.HasError() {
			return errs
		}

		results := filter(objects, findFunc)

		if len(results) > 1 {
			return diag.Errorf("Your query returned more than one result (found %d entries). Please try a more "+
				"specific search criteria.", len(results))
		}
		if len(results) < 1 {
			return diag.Errorf("Your query returned no results. " +
				"Please change your search criteria and try again.")
		}

		result := &results[0]
		log.Printf("[DEBUG] Single %s found", objectName)
		log.Printf("[DEBUG] %+v", *result)
		return setter(d, result)

	}
}

func datasourceBrightboxRecentRead[O brightbox.CreateDated](
	reader func(*brightbox.Client, context.Context) ([]O, error),
	objectName string,
	setter func(*schema.ResourceData, *O) diag.Diagnostics,
	finderGenerator func(*schema.ResourceData) (func(O) bool, diag.Diagnostics),
) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		client := meta.(*CompositeClient).APIClient

		log.Printf("[DEBUG] %s data read called. Retrieving object list", objectName)

		objects, err := reader(client, ctx)
		if err != nil {
			return diag.FromErr(err)
		}

		findFunc, errs := finderGenerator(d)
		if errs.HasError() {
			return errs
		}

		results := filter(objects, findFunc)

		var result *O

		if len(results) > 1 {
			recent, ok := d.GetOk("most_recent")
			if !ok || !recent.(bool) {
				return diag.Errorf("Your query returned more than one result (found %d entries). Please try a more "+
					"specific search criteria.", len(results))
			}
			log.Printf("[DEBUG] Multiple results found and `most_recent` is set")
			result = mostRecent(results)
		} else if len(results) < 1 {
			return diag.Errorf("Your query returned no results. " +
				"Please change your search criteria and try again.")
		} else {
			result = &results[0]
		}
		log.Printf("[DEBUG] Single %s found", objectName)
		log.Printf("[DEBUG] %+v", *result)
		return setter(d, result)

	}
}

func resourceBrightboxUpdate[O, I any](
	putter func(*brightbox.Client, context.Context, I) (*O, error),
	objectName string,
	newFromID func(string) *I,
	updater func(*schema.ResourceData, *I) diag.Diagnostics,
	setter func(*schema.ResourceData, *O) diag.Diagnostics,
) schema.UpdateContextFunc {
	return resourceBrightboxUpdateWithLock(
		putter, objectName, newFromID, updater, setter, nil,
	)
}

func resourceBrightboxUpdateWithLock[O, I any](
	putter func(*brightbox.Client, context.Context, I) (*O, error),
	objectName string,
	newFromID func(string) *I,
	updater func(*schema.ResourceData, *I) diag.Diagnostics,
	setter func(*schema.ResourceData, *O) diag.Diagnostics,
	locksetter schema.UpdateContextFunc,
) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		client := meta.(*CompositeClient).APIClient

		objectOpts := newFromID(d.Id())
		errs := updater(d, objectOpts)
		if errs.HasError() {
			return errs
		}
		log.Printf("[DEBUG] %s update configuration: %+v", objectName, objectOpts)

		object, err := putter(client, ctx, *objectOpts)
		if err != nil {
			return diag.FromErr(err)
		}
		log.Printf("[DEBUG] setting details from returned object: %+v", *object)
		if locksetter != nil && d.HasChange("locked") {
			return locksetter(ctx, d, meta)
		}
		return setter(d, object)
	}
}

func resourceBrightboxDelete[O any](
	deleter func(*brightbox.Client, context.Context, string) (*O, error),
	objectName string,
) schema.DeleteContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		client := meta.(*CompositeClient).APIClient

		log.Printf("[INFO] Deleting %s %s", objectName, d.Id())
		_, err := deleter(client, ctx, d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
		log.Printf("[DEBUG] Deleted cleanly")
		return nil
	}
}

func resourceBrightboxDeleteAndWait[O any](
	deleter func(*brightbox.Client, context.Context, string) (*O, error),
	objectName string,
	pending []string,
	target []string,
	refresher func(*brightbox.Client, context.Context, string) retry.StateRefreshFunc,
) schema.DeleteContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		diags := resourceBrightboxDelete(deleter, objectName)(ctx, d, meta)
		if diags.HasError() {
			return diags
		}

		client := meta.(*CompositeClient).APIClient
		stateConf := retry.StateChangeConf{
			Pending:    pending,
			Target:     target,
			Refresh:    refresher(client, ctx, d.Id()),
			Timeout:    d.Timeout(schema.TimeoutDelete),
			Delay:      checkDelay,
			MinTimeout: minimumRefreshWait,
		}
		_, err := stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId("")
		return nil
	}
}

func resourceBrightboxSetLockState[O any](
	locker func(*brightbox.Client, context.Context, string) (*O, error),
	unlocker func(*brightbox.Client, context.Context, string) (*O, error),
	setter func(*schema.ResourceData, *O) diag.Diagnostics,
) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{},
	) diag.Diagnostics {
		locked := d.Get("locked").(bool)
		log.Printf("[INFO] Setting lock state to %v", locked)
		client := meta.(*CompositeClient).APIClient
		var object *O
		var err error
		if locked {
			object, err = locker(client, ctx, d.Id())
		} else {
			object, err = unlocker(client, ctx, d.Id())
		}
		if err != nil {
			return diag.FromErr(err)
		}
		log.Printf("[DEBUG] setting details from returned object: %+v", *object)
		return setter(d, object)
	}
}
