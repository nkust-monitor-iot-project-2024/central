// Code generated by ent, DO NOT EDIT.

package movement

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/nkust-monitor-iot-project-2024/central/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.Movement {
	return predicate.Movement(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.Movement {
	return predicate.Movement(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.Movement {
	return predicate.Movement(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.Movement {
	return predicate.Movement(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.Movement {
	return predicate.Movement(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.Movement {
	return predicate.Movement(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.Movement {
	return predicate.Movement(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.Movement {
	return predicate.Movement(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.Movement {
	return predicate.Movement(sql.FieldLTE(FieldID, id))
}

// Picture applies equality check predicate on the "picture" field. It's identical to PictureEQ.
func Picture(v []byte) predicate.Movement {
	return predicate.Movement(sql.FieldEQ(FieldPicture, v))
}

// PictureMime applies equality check predicate on the "picture_mime" field. It's identical to PictureMimeEQ.
func PictureMime(v string) predicate.Movement {
	return predicate.Movement(sql.FieldEQ(FieldPictureMime, v))
}

// PictureEQ applies the EQ predicate on the "picture" field.
func PictureEQ(v []byte) predicate.Movement {
	return predicate.Movement(sql.FieldEQ(FieldPicture, v))
}

// PictureNEQ applies the NEQ predicate on the "picture" field.
func PictureNEQ(v []byte) predicate.Movement {
	return predicate.Movement(sql.FieldNEQ(FieldPicture, v))
}

// PictureIn applies the In predicate on the "picture" field.
func PictureIn(vs ...[]byte) predicate.Movement {
	return predicate.Movement(sql.FieldIn(FieldPicture, vs...))
}

// PictureNotIn applies the NotIn predicate on the "picture" field.
func PictureNotIn(vs ...[]byte) predicate.Movement {
	return predicate.Movement(sql.FieldNotIn(FieldPicture, vs...))
}

// PictureGT applies the GT predicate on the "picture" field.
func PictureGT(v []byte) predicate.Movement {
	return predicate.Movement(sql.FieldGT(FieldPicture, v))
}

// PictureGTE applies the GTE predicate on the "picture" field.
func PictureGTE(v []byte) predicate.Movement {
	return predicate.Movement(sql.FieldGTE(FieldPicture, v))
}

// PictureLT applies the LT predicate on the "picture" field.
func PictureLT(v []byte) predicate.Movement {
	return predicate.Movement(sql.FieldLT(FieldPicture, v))
}

// PictureLTE applies the LTE predicate on the "picture" field.
func PictureLTE(v []byte) predicate.Movement {
	return predicate.Movement(sql.FieldLTE(FieldPicture, v))
}

// PictureMimeEQ applies the EQ predicate on the "picture_mime" field.
func PictureMimeEQ(v string) predicate.Movement {
	return predicate.Movement(sql.FieldEQ(FieldPictureMime, v))
}

// PictureMimeNEQ applies the NEQ predicate on the "picture_mime" field.
func PictureMimeNEQ(v string) predicate.Movement {
	return predicate.Movement(sql.FieldNEQ(FieldPictureMime, v))
}

// PictureMimeIn applies the In predicate on the "picture_mime" field.
func PictureMimeIn(vs ...string) predicate.Movement {
	return predicate.Movement(sql.FieldIn(FieldPictureMime, vs...))
}

// PictureMimeNotIn applies the NotIn predicate on the "picture_mime" field.
func PictureMimeNotIn(vs ...string) predicate.Movement {
	return predicate.Movement(sql.FieldNotIn(FieldPictureMime, vs...))
}

// PictureMimeGT applies the GT predicate on the "picture_mime" field.
func PictureMimeGT(v string) predicate.Movement {
	return predicate.Movement(sql.FieldGT(FieldPictureMime, v))
}

// PictureMimeGTE applies the GTE predicate on the "picture_mime" field.
func PictureMimeGTE(v string) predicate.Movement {
	return predicate.Movement(sql.FieldGTE(FieldPictureMime, v))
}

// PictureMimeLT applies the LT predicate on the "picture_mime" field.
func PictureMimeLT(v string) predicate.Movement {
	return predicate.Movement(sql.FieldLT(FieldPictureMime, v))
}

// PictureMimeLTE applies the LTE predicate on the "picture_mime" field.
func PictureMimeLTE(v string) predicate.Movement {
	return predicate.Movement(sql.FieldLTE(FieldPictureMime, v))
}

// PictureMimeContains applies the Contains predicate on the "picture_mime" field.
func PictureMimeContains(v string) predicate.Movement {
	return predicate.Movement(sql.FieldContains(FieldPictureMime, v))
}

// PictureMimeHasPrefix applies the HasPrefix predicate on the "picture_mime" field.
func PictureMimeHasPrefix(v string) predicate.Movement {
	return predicate.Movement(sql.FieldHasPrefix(FieldPictureMime, v))
}

// PictureMimeHasSuffix applies the HasSuffix predicate on the "picture_mime" field.
func PictureMimeHasSuffix(v string) predicate.Movement {
	return predicate.Movement(sql.FieldHasSuffix(FieldPictureMime, v))
}

// PictureMimeEqualFold applies the EqualFold predicate on the "picture_mime" field.
func PictureMimeEqualFold(v string) predicate.Movement {
	return predicate.Movement(sql.FieldEqualFold(FieldPictureMime, v))
}

// PictureMimeContainsFold applies the ContainsFold predicate on the "picture_mime" field.
func PictureMimeContainsFold(v string) predicate.Movement {
	return predicate.Movement(sql.FieldContainsFold(FieldPictureMime, v))
}

// HasEvent applies the HasEdge predicate on the "event" edge.
func HasEvent() predicate.Movement {
	return predicate.Movement(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, EventTable, EventPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasEventWith applies the HasEdge predicate on the "event" edge with a given conditions (other predicates).
func HasEventWith(preds ...predicate.Event) predicate.Movement {
	return predicate.Movement(func(s *sql.Selector) {
		step := newEventStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Movement) predicate.Movement {
	return predicate.Movement(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Movement) predicate.Movement {
	return predicate.Movement(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Movement) predicate.Movement {
	return predicate.Movement(sql.NotPredicates(p))
}
