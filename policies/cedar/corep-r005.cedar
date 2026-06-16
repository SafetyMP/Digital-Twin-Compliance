// COREP-R005: Capital adjustments require CapitalManager role.
@id("COREP-R005")
@description("Capital adjustment entries require role=CapitalManager")
permit(
  principal,
  action == DigitalTwin::Action::"adjust",
  resource
)
when {
  resource is DigitalTwin::CapitalAdjustment &&
  principal in DigitalTwin::Role::"CapitalManager"
};
