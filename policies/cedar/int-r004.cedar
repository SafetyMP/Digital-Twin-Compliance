// INT-R004: Payment amounts above €500K require Approver role.
@id("INT-R004")
@description("Payment amount greater than 500000 EUR requires role=Approver")
permit(
  principal,
  action == DigitalTwin::Action::"approve",
  resource
)
when {
  resource is DigitalTwin::Payment &&
  (
    resource.amountEur <= 500000 ||
    principal in DigitalTwin::Role::"Approver"
  )
};
