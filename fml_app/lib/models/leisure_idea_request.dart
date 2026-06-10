class LeisureIdeaRequest {
  final String location;
  final String activityLevel;
  final String budget;
  final String extraContext;

  LeisureIdeaRequest({
    required this.location,
    required this.activityLevel,
    required this.budget,
    required this.extraContext,
  });

  Map<String, dynamic> toJson() {
    return {
      'location': location,
      'activity_level': activityLevel,
      'budget': budget,
      'extra_context': extraContext,
    };
  }
}