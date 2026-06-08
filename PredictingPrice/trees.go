package predictingprice

type XGBModel struct {
	Learner struct {
		LearnerModelParam struct {
			BaseScore float64 `json:"base_score"`
		} `json:"learner_model_param"`
		FeatureNames    []string `json:"feature_names"`
		GradientBooster struct {
			Model struct {
				Trees []Tree `json:"trees"`
			} `json:"model"`
		} `json:"gradient_booster"`
	} `json:"learner"`
}

type Tree struct {
	BaseWeights     []float64 `json:"base_weights"`
	LeftChildren    []int     `json:"left_children"`
	RightChildren   []int     `json:"right_children"`
	SplitConditions []float64 `json:"split_conditions"`
	SplitIndices    []int     `json:"split_indices"`
	SplitType       []int     `json:"split_type"`
	
	// Categorical Metadata
	CategoriesNodes []int `json:"categories_nodes"`
	CategoriesSeg   []int `json:"categories_segments"`
	CategoriesSizes []int `json:"categories_sizes"`
	Categories      []int `json:"categories"`
}

// Predict is now a method of XGBModel
func (m *XGBModel) Predict(features []float32) float64 {
	sum := m.Learner.LearnerModelParam.BaseScore

	for _, tree := range m.Learner.GradientBooster.Model.Trees {
		sum += tree.Evaluate(features, 0)
	}
	return sum
}

// Evaluate is now a method of Tree
func (t *Tree) Evaluate(features []float32, nodeIdx int) float64 {
	// Leaf node check
	if t.LeftChildren[nodeIdx] == -1 {
		return t.BaseWeights[nodeIdx]
	}

	// Categorical Split Logic (SplitType == 1)
	if len(t.SplitType) > nodeIdx && t.SplitType[nodeIdx] == 1 {
		segStart, segSize := -1, 0
		for i, n := range t.CategoriesNodes {
			if n == nodeIdx {
				segStart = t.CategoriesSeg[i]
				segSize = t.CategoriesSizes[i]
				break
			}
		}

		featIdx := t.SplitIndices[nodeIdx]
		featVal := int(features[featIdx])

		isLeft := false
		// Check against the mask
		for i := 0; i < segSize; i++ {
			if t.Categories[segStart+i] == featVal {
				isLeft = true
				break
			}
		}

		if isLeft {
			return t.Evaluate(features, t.LeftChildren[nodeIdx])
		}
		return t.Evaluate(features, t.RightChildren[nodeIdx])
	}

	// Numerical Split (SplitType == 0)
	featIdx := t.SplitIndices[nodeIdx]
	if features[featIdx] < float32(t.SplitConditions[nodeIdx]) {
		return t.Evaluate(features, t.LeftChildren[nodeIdx])
	}
	return t.Evaluate(features, t.RightChildren[nodeIdx])
}