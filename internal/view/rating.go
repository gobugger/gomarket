package view

type Rating struct {
	Whole    int
	Half     bool
	Real     float64
	Positive float64
}

func calculateRating(reviews []ProductReview) Rating {
	var rating = 0.0
	var positive = 0.0

	for _, review := range reviews {
		rating += float64(review.Review.Grade)
		if review.Review.Grade > 2 {
			positive += 1.0
		}
	}

	if len(reviews) > 0 {
		rating /= float64(len(reviews))
		positive /= float64(len(reviews))
		positive *= 100
	}

	half := false
	whole := int(rating)
	if rating-float64(whole) > 0.75 {
		whole += 1
	} else if rating-float64(whole) > 0.25 {
		half = true
	}

	return Rating{
		Whole:    whole,
		Half:     half,
		Real:     rating,
		Positive: positive,
	}
}

func calculateRating2(reviews []Review) Rating {
	var rating = 0.0
	var positive = 0.0

	for _, review := range reviews {
		rating += float64(review.Review.Grade)
		if review.Review.Grade > 2 {
			positive += 1.0
		}
	}

	if len(reviews) > 0 {
		rating /= float64(len(reviews))
		positive /= float64(len(reviews))
		positive *= 100
	}

	half := false
	whole := int(rating)
	if rating-float64(whole) > 0.75 {
		whole += 1
	} else if rating-float64(whole) > 0.25 {
		half = true
	}

	return Rating{
		Whole:    whole,
		Half:     half,
		Real:     rating,
		Positive: positive,
	}
}
